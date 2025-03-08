package downloading

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sync"
	"uv_server/internal/uv_server/business"
	commonJobMessages "uv_server/internal/uv_server/business/common_job_messages"
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

	job_in chan interface{}
}

func NewDownloadingWf(
	uuid string,
	config *config.Config,
	jobCtx context.Context,
	job_in chan interface{},
	job_out chan interface{},
) *DownloadingWf {
	object := &DownloadingWf{}

	object.uuid = uuid
	object.log = loggers.BusinessLogger.WithFields(
		logrus.Fields{
			"component": "DownloadingWf",
			"uuid":      uuid})
	object.config = config

	object.jobCtx = jobCtx

	object.job_in = job_in
	_ = job_out

	return object
}

var youtubeRegex = regexp.MustCompile(`(?:https?:\/\/)?(?:www\.|m\.)?(?:youtube\.com\/(?:watch\?v=|embed\/|v\/|shorts\/)|youtu\.be\/)([a-zA-Z0-9_-]{11})`)

func (w *DownloadingWf) Run(wg *sync.WaitGroup, request *jobmessages.Request) {
	defer wg.Done()

	url := *request.Url
	w.log.Tracef("serving downloading request for url: %v", url)

	source, err := w.getSourceFromUrl(url)
	if err != nil {
		w.log.Error(err)
		w.job_in <- commonJobMessages.Error{Reason: err.Error()}
		return
	}

	if source == business.Youtube {
		err := w.startDownloadingFromYoutube(url)
		if err != nil {
			w.log.Errorf("start downloading from youtube failed with error: %v", err)
			w.job_in <- commonJobMessages.Error{Reason: err.Error()}
		}
	} else {
		w.log.Fatalf("downloading for %v is not implemented", source)
	}

	for {
		select {
		case <-w.jobCtx.Done():
			w.log.Debugf("workflow cancelled: %v", w.jobCtx.Err().Error())

			switch w.jobCtx.Err() {
			case context.DeadlineExceeded:
				w.job_in <- commonJobMessages.Error{Reason: "Timeout exceeded"}
			case context.Canceled:
				w.job_in <- commonJobMessages.Error{Reason: "Workflow cancelled"}
			}

			return
		}
	}
}

func isYoutube(url string) bool {
	return youtubeRegex.MatchString(url)
}

func (w *DownloadingWf) getSourceFromUrl(url string) (business.Source, error) {
	if isYoutube(url) {
		return business.Youtube, nil
	}

	return business.Unknown, fmt.Errorf("unable to idenitify source of the url: %v", url)
}

func normalizeYoutubeUrl(url string) (string, error) {
	if !isYoutube(url) {
		return "", errors.New("invalid YouTube URL")
	}

	matches := youtubeRegex.FindStringSubmatch(url)
	return fmt.Sprintf("https://www.youtube.com/watch?v=%s", matches[1]), nil
}

func (w *DownloadingWf) startDownloadingFromYoutube(url string) error {
	log := w.log.WithField("source", "youtube")

	log.Debugf("starting downloading")
	url, err := normalizeYoutubeUrl(url)
	if err != nil {
		return err
	}

	log.Debugf("normalized url is: %v", url)

	return nil
}
