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
	jobIn  chan<- interface{}

	downloaderOut <-chan interface{}
	downloader    wfData.Downloader

	database data.Database

	fileId int64

	startDownloading func(
		downloaderWg *sync.WaitGroup,
		url string,
	) error

	startDownloadingFromYoutube func(
		downloaderWg *sync.WaitGroup,
		url string,
	) error
}

func NewDownloadingWf(
	uuid string,
	config *config.Config,
	jobCtx context.Context,
	jobIn chan<- interface{},
	job_out <-chan interface{},
	downloader wfData.Downloader,
	downloaderOut <-chan interface{},
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

	object.jobIn = jobIn
	_ = job_out

	object.downloader = downloader
	object.downloaderOut = downloaderOut

	object.database = database

	object.injectInternalDependencies()

	return object
}

func (w *DownloadingWf) injectInternalDependencies() {
	w.startDownloading = func(
		downloaderWg *sync.WaitGroup,
		url string,
	) error {
		return startDownloading(w, downloaderWg, url)
	}

	w.startDownloadingFromYoutube = func(
		downloaderWg *sync.WaitGroup,
		url string,
	) error {
		return startDownloadingFromYoutube(w, downloaderWg, url)
	}
}

var youtubeRegex = regexp.MustCompile(`(?:https?:\/\/)?(?:www\.|m\.)?(?:youtube\.com\/(?:watch\?v=|embed\/|v\/|shorts\/)|youtu\.be\/)([a-zA-Z0-9_-]{11})`)

func (w *DownloadingWf) Run(wg *sync.WaitGroup, request *jobmessages.Request) {
	defer wg.Done()

	url := *request.Url
	w.log.Tracef("serving downloading request for url: %v", url)

	if len(url) == 0 {
		errMsg := "url is empty"
		w.log.Errorf(errMsg)
		w.jobIn <- &cjmessages.Error{Reason: errMsg}
		return
	}

	var downloaderWg sync.WaitGroup
	err := w.startDownloading(&downloaderWg, url)
	if err != nil {
		w.log.Errorf("start downloading failed with error: %v", err)
		w.jobIn <- &cjmessages.Error{Reason: err.Error()}
		return
	}

	w.jobIn <- &jobmessages.Progress{Percentage: 0}
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
				w.jobIn <- &cjmessages.Error{Reason: "Timeout exceeded"}
			case context.Canceled:
				w.jobIn <- &cjmessages.Error{Reason: "Workflow cancelled"}
			}

			return
		case msg := <-w.downloaderOut:
			if tMsg, ok := msg.(*wfData.Progress); ok {
				now := time.Now()
				if now.Sub(lastProgressTs).Seconds() > 0 {
					w.jobIn <- &jobmessages.Progress{Percentage: tMsg.Percentage}
					lastProgressTs = now
				}
			} else if _, ok := msg.(*wfData.Error); ok {
				downloaderWg.Wait()
				err := w.database.DeleteFile(&data.File{Id: w.fileId})
				if err != nil {
					w.log.Fatalf(
						"failed to delete file with id %v, error is %v",
						w.fileId, err)
				}

				w.jobIn <- &cjmessages.Error{Reason: "downloading failed"}
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
				w.jobIn <- &jobmessages.Progress{Percentage: 100}
				w.jobIn <- &cjmessages.Done{}
				return
			} else {
				downloaderWg.Wait()
				w.jobIn <- &cjmessages.InternalError
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

	w.log.Errorf("unable to idenitify source of the url: %v", url)

	return data.Unknown, fmt.Errorf("unable to identify source")
}

func normalizeYoutubeUrl(url string) (string, error) {
	if !isYoutube(url) {
		return "", errors.New("invalid YouTube URL")
	}

	matches := youtubeRegex.FindStringSubmatch(url)
	return fmt.Sprintf("https://www.youtube.com/watch?v=%s", matches[1]), nil
}

func startDownloading(
	w *DownloadingWf,
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

func startDownloadingFromYoutube(
	w *DownloadingWf,
	downloaderWg *sync.WaitGroup,
	url string,
) error {
	log := w.log.WithField("source", "youtube")

	log.Debugf("starting downloading")

	downloaderWg.Add(1)
	go w.downloader.Download(downloaderWg, url)

	return nil
}
