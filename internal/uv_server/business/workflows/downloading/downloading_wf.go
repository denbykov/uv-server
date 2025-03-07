package downloading

import (
	"context"
	"fmt"
	"regexp"
	"uv_server/internal/uv_server/business"
	jobmessages "uv_server/internal/uv_server/business/workflows/downloading/job_messages"
	"uv_server/internal/uv_server/common/loggers"
	"uv_server/internal/uv_server/config"

	"github.com/sirupsen/logrus"
)

type DownloadingWf struct {
	uuid string

	log    *logrus.Entry
	config *config.Config

	ctx    context.Context
	cancel context.CancelFunc
}

func NewDownloadingWf(
	uuid string,
	config *config.Config,
	ctx context.Context,
	cancel context.CancelFunc,
) *DownloadingWf {
	object := &DownloadingWf{}

	object.uuid = uuid
	object.log = loggers.BusinessLogger.WithFields(
		logrus.Fields{
			"component": "DownloadingWf",
			"uuid":      uuid})
	object.config = config
	object.ctx = ctx
	object.cancel = cancel

	return object
}

func (w *DownloadingWf) Run(request *jobmessages.Request) {
	url := *request.Url
	w.log.Tracef("serving downloading request for url: %v", url)

	source, err := w.getSourceFromUrl(url)

	if err != nil {
		w.log.Error(err)
		return
	}

	w.log.Tracef("the soucre is: %v", source)

	for {
		select {
		case <-w.ctx.Done():
			w.log.Debug("I'm done")
			return
		}
	}
}

func isYoutube(url string) bool {
	var regex = regexp.MustCompile(`(?:https?:\/\/)?(?:www\.|m\.)?(?:youtube\.com\/(?:watch\?v=|embed\/|v\/|shorts\/)|youtu\.be\/)([a-zA-Z0-9_-]{11})`)
	return regex.MatchString(url)
}

func (w *DownloadingWf) getSourceFromUrl(url string) (business.Source, error) {
	if isYoutube(url) {
		return business.Youtube, nil
	}

	return business.Unknown, fmt.Errorf("unable to idenitify source of the url: %v", url)
}
