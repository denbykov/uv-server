package downloading

import (
	"context"
	"sync"
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

func (w *DownloadingWf) Run(wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-w.ctx.Done():
			w.log.Debug("I'm done")
			return
		}
	}
}
