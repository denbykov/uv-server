package downloaders

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"reflect"
	"sync"

	"github.com/sirupsen/logrus"

	businessData "uv_server/internal/uv_server/business/workflows/downloading/data"
	"uv_server/internal/uv_server/common/loggers"
	"uv_server/internal/uv_server/config"
)

type mtype int

const (
	DownloadingProgress mtype = 1
	DownloadingDone     mtype = 2
	DownloadingFailed   mtype = 3
)

type YtDownloader struct {
	log    *logrus.Entry
	config *config.Config

	jobCtx context.Context

	wf_out chan interface{}

	child_out chan interface{}
}

func NewYtDownloader(
	config *config.Config,
	jobCtx context.Context,
	wf_out chan interface{},
) *YtDownloader {
	object := &YtDownloader{}
	object.log = loggers.DataLogger.WithField("component", "YtDownloader")
	object.config = config

	object.jobCtx = jobCtx

	object.wf_out = wf_out

	object.child_out = make(chan interface{}, 1)

	return object
}

func (d *YtDownloader) handleProgressMessage(
	message map[string]interface{},
) error {
	d.log.Tracef("Handling progress message: %v", message)

	percentage, ok := message["percentage"]

	if !ok {
		return errors.New("progress message does not contain " +
			"a \"percentage\" field")
	}

	businessMessage := &businessData.Progress{
		Percentage: percentage.(float64),
	}

	d.child_out <- businessMessage
	return nil
}

func (d *YtDownloader) handleDoneMessage(
	message map[string]interface{},
) error {
	d.log.Tracef("Handling done message: %v", message)

	filename, ok := message["filename"]

	if !ok {
		return errors.New("done message does not contain " +
			"a \"filename\" field")
	}

	businessMessage := &businessData.Done{
		Filename: filename.(string),
	}

	d.child_out <- businessMessage
	return nil
}

func (d *YtDownloader) Download(wg *sync.WaitGroup, url string) {
	d.log.Debugf("downloading file from url: %v", url)

	defer wg.Done()

	wd, err := os.Getwd()
	if err != nil {
		d.log.Fatal(err)
	}

	process, stdout, err := d.startProcess(wd, url)
	if err != nil {
		d.log.Fatal(err)
	}

	go d.listenToChild(stdout)

	for {
		select {
		case <-d.jobCtx.Done():
			stdout.Close()
			d.cleanUp(process, false)
			return
		case msg := <-d.child_out:
			if typedMsg, ok := msg.(businessData.Progress); ok {
				d.wf_out <- typedMsg
			} else if typedMsg, ok := msg.(businessData.Done); ok {
				d.wf_out <- typedMsg
				d.cleanUp(process, true)
				return
			} else if typedMsg, ok := msg.(businessData.Error); ok {
				d.wf_out <- typedMsg
				d.cleanUp(process, false)
				return
			} else {
				d.log.Fatalf("Unknown message type: %v", reflect.TypeOf(msg))
			}
		}
	}
}

func (d *YtDownloader) cleanUp(process *exec.Cmd, gracefulExit bool) {
	if !gracefulExit {
		err := process.Process.Kill()
		if err != nil {
			d.log.Fatal(err)
		}
	}

	err := process.Wait()
	if err != nil {
		d.log.Fatal(err)
	}
}

func (d *YtDownloader) startProcess(wd string, url string) (*exec.Cmd, io.ReadCloser, error) {
	executable := path.Join(wd, d.config.ScriptsLocation, "downloader")
	storage_location := "."

	process := exec.Command(
		executable,
		"--url", url,
		"--dir", storage_location,
		"--ffmpeg_location", d.config.FfmpegLocation)

	stdout, err := process.StdoutPipe()
	if err != nil {
		d.log.Fatal(err)
	}

	err = process.Start()
	if err != nil {
		d.log.Fatal(err)
	}

	return process, stdout, nil
}

func (d *YtDownloader) listenToChild(stdout io.ReadCloser) {
	reader := bufio.NewReader(stdout)

	for {
		message, err := reader.ReadBytes('\n')
		d.log.Tracef("Handling script message: %v", string(message))

		if err != nil {
			d.log.Errorf("failed to read message from script: %v", err)
			d.child_out <- businessData.Error{Reason: "downloading failed"}
			return
		}

		parsedMessage, t, err := parseChildMessage(message)

		if err != nil {
			d.log.Error(err)
			d.child_out <- businessData.Error{Reason: "downloading failed"}
			return
		}

		switch mtype(t.(float64)) {
		case DownloadingProgress:
			err := d.handleProgressMessage(parsedMessage)
			if err != nil {
				d.log.Errorf("failed to handle progress message: %v, reason: %v",
					parsedMessage, err)

				d.child_out <- businessData.Error{Reason: "downloading failed"}
				return
			}
		case DownloadingDone:
			err := d.handleDoneMessage(parsedMessage)

			if err != nil {
				d.log.Errorf("failed to handle done message: %v, reason: %v",
					parsedMessage, err)

				d.child_out <- businessData.Error{Reason: "downloading failed"}
				return
			}
			return
		default:
			d.log.Errorf("no message handler for type: %v", err)
			d.child_out <- businessData.Error{Reason: "downloading failed"}
			return
		}
	}
}

func parseChildMessage(message []byte) (map[string]interface{}, interface{}, error) {
	var parsedMessage map[string]interface{}
	err := json.Unmarshal(message, &parsedMessage)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal message from script: %v", err)
	}

	t, ok := parsedMessage["type"]
	if !ok {
		return nil, nil, fmt.Errorf("message from does not contain \"type\" field")
	}

	return parsedMessage, t, nil
}
