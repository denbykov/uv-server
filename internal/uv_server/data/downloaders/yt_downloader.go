package downloaders

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"reflect"
	"strings"
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
	uuid string

	log    *logrus.Entry
	config *config.Config

	jobCtx context.Context

	wf_out chan<- interface{}

	child_out chan interface{}

	to_clean chan<- string
}

func NewYtDownloader(
	uuid string,
	config *config.Config,
	jobCtx context.Context,
	wf_out chan<- interface{},
	to_clean chan<- string,
) *YtDownloader {
	object := &YtDownloader{}
	object.log = loggers.DataLogger.WithFields(
		logrus.Fields{
			"component": "DownloadingWf",
			"uuid":      uuid},
	)

	object.uuid = uuid

	object.config = config
	object.jobCtx = jobCtx
	object.wf_out = wf_out
	object.child_out = make(chan interface{}, 1)
	object.to_clean = to_clean

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

func (d *YtDownloader) handleFailedMessage(
	message map[string]interface{},
) error {
	d.log.Tracef("Handling failed message: %v", message)

	msg, ok := message["msg"]

	if !ok {
		return errors.New("failed message does not contain " +
			"a \"msg\" field")
	}

	businessMessage := &businessData.Error{
		Reason: msg.(string),
	}

	d.child_out <- businessMessage
	return nil
}

func copyFile(src, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destinationStat, err := os.Stat(dst)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		destination, err := os.Create(dst)
		if err != nil {
			return err
		}

		defer destination.Close()
		_, err = io.Copy(destination, source)
		return err
	}

	if err != nil {
		return err
	}

	if !destinationStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}

	defer destination.Close()
	_, err = io.Copy(destination, source)
	return err
}

func (d *YtDownloader) ensureDirectoryExists(dir string) {
	stat, err := os.Stat(dir)

	if err != nil && errors.Is(err, os.ErrNotExist) {
		if err := os.Mkdir(dir, os.ModePerm); err != nil {
			log.Fatal(err)
		}

		return
	}

	if err != nil {
		log.Fatalf("failed to stat temp dir: %v", err)
	}

	if !stat.Mode().IsDir() {
		log.Fatalf("temp dir path is taken by non-direcotry: %v", err)
	}
}

func (d *YtDownloader) Download(wg *sync.WaitGroup, url string, storageDir string) {
	d.log.Debugf("downloading file from url: %v", url)

	defer wg.Done()

	wd, err := os.Getwd()
	if err != nil {
		d.log.Fatal(err)
	}

	tempDir := path.Join(wd, "tmp", d.uuid)
	d.ensureDirectoryExists(tempDir)

	process, stdout, err := d.startProcess(wd, url, tempDir)
	if err != nil {
		d.log.Fatal(err)
	}

	var childWg sync.WaitGroup
	childWg.Add(1)
	go d.listenToChild(&childWg, stdout)

	for {
		select {
		case <-d.jobCtx.Done():
			stdout.Close()
			d.cleanUp(process, &childWg, false, tempDir)
			return
		case msg := <-d.child_out:
			if typedMsg, ok := msg.(*businessData.Progress); ok {
				d.wf_out <- typedMsg
			} else if typedMsg, ok := msg.(*businessData.Done); ok {
				sfn := strings.Split(typedMsg.Filename, string(os.PathSeparator))
				typedMsg.Filename = sfn[len(sfn)-1]

				err := copyFile(
					path.Join(tempDir, typedMsg.Filename),
					path.Join(storageDir, typedMsg.Filename),
				)

				if err != nil {
					d.log.Fatalf("Failed to copy file: %v", err)
				}

				d.cleanUp(process, &childWg, true, tempDir)
				d.wf_out <- typedMsg
				return
			} else if typedMsg, ok := msg.(*businessData.Error); ok {
				d.wf_out <- typedMsg
				d.cleanUp(process, &childWg, false, tempDir)
				return
			} else {
				d.log.Fatalf("Unknown message type: %v", reflect.TypeOf(msg))
			}
		}
	}
}

func (d *YtDownloader) cleanUp(
	process *exec.Cmd,
	childWg *sync.WaitGroup,
	gracefulExit bool,
	tempDir string,
) {
	d.log.WithField("graceful", gracefulExit).Trace("Cleaning up")

	childWg.Wait()

	if !gracefulExit {
		err := process.Process.Kill()
		if err != nil {
			d.log.Error(err)
		}
	}

	err := process.Wait()
	if err != nil {
		d.log.Tracef("downlaoder executable exited with: %v", err)
	}

	d.to_clean <- tempDir

	d.log.Trace("Done cleaning up")
}

func (d *YtDownloader) startProcess(wd string, url string, dir string) (*exec.Cmd, io.ReadCloser, error) {
	executable := path.Join(wd, d.config.ToolsLocation, "downloader")

	process := exec.Command(
		executable,
		"--url", url,
		"--dir", dir,
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

func (d *YtDownloader) listenToChild(wg *sync.WaitGroup, stdout io.ReadCloser) {
	defer wg.Done()

	reader := bufio.NewReader(stdout)

	for {
		message, err := reader.ReadBytes('\n')
		d.log.Tracef("Handling script message: %v", string(message))

		if err != nil && !errors.Is(err, os.ErrClosed) {
			d.log.Errorf("failed to read message from script: %v", err)
			d.child_out <- &businessData.Error{Reason: "downloading failed"}
			return
		}

		if err != nil && errors.Is(err, os.ErrClosed) {
			d.log.Tracef("Child output closed")
			return
		}

		parsedMessage, t, err := parseChildMessage(message)

		if err != nil {
			d.log.Error(err)
			d.child_out <- &businessData.Error{Reason: "downloading failed"}
			return
		}

		switch mtype(t.(float64)) {
		case DownloadingProgress:
			err := d.handleProgressMessage(parsedMessage)
			if err != nil {
				d.log.Errorf("failed to handle progress message: %v, reason: %v",
					parsedMessage, err)

				d.child_out <- &businessData.Error{Reason: "downloading failed"}
				return
			}
		case DownloadingDone:
			err := d.handleDoneMessage(parsedMessage)

			if err != nil {
				d.log.Errorf("failed to handle done message: %v, reason: %v",
					parsedMessage, err)

				d.child_out <- &businessData.Error{Reason: "downloading failed"}
				return
			}

			return
		case DownloadingFailed:
			err := d.handleFailedMessage(parsedMessage)

			if err != nil {
				d.log.Errorf("failed to handle error message: %v, reason: %v",
					parsedMessage, err)

				d.child_out <- &businessData.Error{Reason: "downloading failed"}
				return
			}
			return
		default:
			d.log.Errorf("no message handler for type: %v", err)
			d.child_out <- &businessData.Error{Reason: "downloading failed"}
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
