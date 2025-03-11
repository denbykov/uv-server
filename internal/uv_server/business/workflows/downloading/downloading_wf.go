package downloading

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"sync"
	"time"
	cjmessages "uv_server/internal/uv_server/business/common_job_messages"
	"uv_server/internal/uv_server/business/data"
	wfData "uv_server/internal/uv_server/business/workflows/downloading/data"
	jobmessages "uv_server/internal/uv_server/business/workflows/downloading/job_messages"
	"uv_server/internal/uv_server/common/loggers"
	"uv_server/internal/uv_server/config"

	"github.com/sirupsen/logrus"
)

type DownloadingWf struct {
	uuid string

	log    *logrus.Entry
	config *config.Config

	jobCtx context.Context
	job_in chan<- interface{}

	downloader_out <-chan interface{}
	downloader     wfData.Downloader

	database data.Database

	fileId int64
}

func NewDownloadingWf(
	uuid string,
	config *config.Config,
	jobCtx context.Context,
	job_in chan<- interface{},
	job_out <-chan interface{},
	downloader wfData.Downloader,
	downloader_out <-chan interface{},
	database data.Database,
) *DownloadingWf {
	object := &DownloadingWf{}

	object.uuid = uuid
	object.log = loggers.BusinessLogger.WithFields(
		logrus.Fields{
			"component": "DownloadingWf",
			"uuid":      uuid},
	)
	object.config = config

	object.jobCtx = jobCtx

	object.job_in = job_in
	_ = job_out

	object.downloader = downloader
	object.downloader_out = downloader_out

	object.database = database

	return object
}

var youtubeRegex = regexp.MustCompile(`(?:https?:\/\/)?(?:www\.|m\.)?(?:youtube\.com\/(?:watch\?v=|embed\/|v\/|shorts\/)|youtu\.be\/)([a-zA-Z0-9_-]{11})`)

func (w *DownloadingWf) Run(wg *sync.WaitGroup, request *jobmessages.Request) {
	defer wg.Done()

	url := *request.Url
	w.log.Tracef("serving downloading request for url: %v", url)

	var downloaderWg sync.WaitGroup
	err := w.startDownloading(&downloaderWg, url)
	if err != nil {
		w.log.Errorf("start downloading failed with error: %v", err)
		w.job_in <- cjmessages.Error{Reason: err.Error()}
		return
	}

	w.job_in <- jobmessages.Progress{Percentage: 0}
	lastProgressTs := time.Now()

	for {
		select {
		case <-w.jobCtx.Done():
			w.log.Debugf("workflow cancelled: %v", w.jobCtx.Err().Error())
			downloaderWg.Wait()

			err := w.database.DeleteFile(&data.File{Id: w.fileId})
			if err != nil {
				w.log.Fatalf(
					"failed to delete file with id %v, error is %v",
					w.fileId, err)
			}

			switch w.jobCtx.Err() {
			case context.DeadlineExceeded:
				w.job_in <- cjmessages.Error{Reason: "Timeout exceeded"}
			case context.Canceled:
				w.job_in <- cjmessages.Error{Reason: "Workflow cancelled"}
			}

			return
		case msg := <-w.downloader_out:
			if tMsg, ok := msg.(*wfData.Progress); ok {
				now := time.Now()
				if now.Sub(lastProgressTs).Seconds() > 0 {
					w.job_in <- jobmessages.Progress{Percentage: tMsg.Percentage}
					lastProgressTs = now
				}
			} else if tMsg, ok := msg.(*wfData.Error); ok {
				w.job_in <- cjmessages.Error{Reason: tMsg.Reason}
				downloaderWg.Wait()
				return
			} else if tMsg, ok := msg.(*wfData.Done); ok {
				file := &data.File{
					Id:     w.fileId,
					Path:   sql.NullString{String: tMsg.Filename, Valid: true},
					Status: data.FsFinished,
				}

				err := w.database.UpdateFilePath(file)
				if err != nil {
					w.log.Fatalf("failed to update path for file with id %v", w.fileId)
				}

				err = w.database.UpdateFileStatus(file)
				if err != nil {
					w.log.Fatalf("failed to update status for file with id %v", w.fileId)
				}

				downloaderWg.Wait()
				w.job_in <- jobmessages.Progress{Percentage: 100}
				w.job_in <- cjmessages.Done{}
				return
			} else {
				w.job_in <- cjmessages.InternalError
				downloaderWg.Wait()
				return
			}
		}
	}
}

func isYoutube(url string) bool {
	return youtubeRegex.MatchString(url)
}

func (w *DownloadingWf) getSourceFromUrl(url string) (data.Source, error) {
	if isYoutube(url) {
		return data.Youtube, nil
	}

	return data.Unknown, fmt.Errorf("unable to idenitify source of the url: %v", url)
}

func normalizeYoutubeUrl(url string) (string, error) {
	if !isYoutube(url) {
		return "", errors.New("invalid YouTube URL")
	}

	matches := youtubeRegex.FindStringSubmatch(url)
	return fmt.Sprintf("https://www.youtube.com/watch?v=%s", matches[1]), nil
}

func (w *DownloadingWf) startDownloading(
	downloaderWg *sync.WaitGroup,
	url string,
) error {
	source, err := w.getSourceFromUrl(url)
	if err != nil {
		return err
	}

	url, err = w.normalizeUrl(url, source)
	if err != nil {
		return err
	}
	w.log.Debugf("normalized url is: %v", url)

	file, err := w.database.GetFileByUrl(url)
	if err != nil {
		w.log.Fatalf("failed to get file by url")
	}

	if file != nil {
		return fmt.Errorf("file already exists")
	}

	if source == data.Youtube {
		err := w.startDownloadingFromYoutube(downloaderWg, url)
		if err != nil {
			return err
		}
	} else {
		w.log.Fatalf("downloading for %v is not implemented", source)
	}

	w.fileId, err = w.database.InsertFile(&data.File{
		SourceUrl: url,
		Source:    source,
		Status:    data.FsDownloading,
	})

	if err != nil {
		w.log.Fatal(err)
	}

	return nil
}

func (w *DownloadingWf) normalizeUrl(
	url string,
	source data.Source,
) (string, error) {
	if source == data.Youtube {
		return normalizeYoutubeUrl(url)
	} else {
		w.log.Fatalf("normalization for %v is not implemented", source)
	}

	return "", nil
}

func (w *DownloadingWf) startDownloadingFromYoutube(
	downloaderWg *sync.WaitGroup,
	url string,
) error {
	log := w.log.WithField("source", "youtube")

	log.Debugf("starting downloading")

	downloaderWg.Add(1)
	go w.downloader.Download(downloaderWg, url)

	return nil
}
