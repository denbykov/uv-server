package data

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path"

	"github.com/sirupsen/logrus"

	business_data "uv_server/internal/uv_server/business/workflows/download/data"
	"uv_server/internal/uv_server/common/loggers"
	"uv_server/internal/uv_server/config"
)

type mtype int

const (
	DownloadingProgress mtype = 1
	DownloadingDone     mtype = 2
	DownloadingFailed   mtype = 3
)

type Downloader struct {
	log        *logrus.Entry
	config     *config.Config
	onProgress func(*business_data.ProgressMessage)
}

func (d *Downloader) RegisterOnProgress(
	onProgress func(*business_data.ProgressMessage)) {
	d.onProgress = onProgress
}

func NewDownloader(config *config.Config) *Downloader {
	object := &Downloader{}
	object.log = loggers.DataLogger.WithField("component", "Downloader")
	object.config = config

	return object
}

func (d *Downloader) handleProgressMessage(
	message map[string]interface{},
) error {
	d.log.Tracef("Handling progress message: %v", message)

	percentage, ok := message["percentage"]

	if !ok {
		return errors.New("progress message does not contain " +
			"a \"percentage\" field")
	}

	businessMessage := &business_data.ProgressMessage{
		Percentage: percentage.(float64),
	}

	d.onProgress(businessMessage)

	return nil
}

func (d *Downloader) handleDoneMessage(
	message map[string]interface{},
) (string, error) {
	d.log.Tracef("Handling done message: %v", message)

	filename, ok := message["filename"]

	if !ok {
		return "", errors.New("done message does not contain " +
			"a \"filename\" field")
	}

	return filename.(string), nil
}

func (d *Downloader) Download(url string) (string, error) {
	d.log.Debugf("Downloading file: %v", url)

	wd, err := os.Getwd()
	if err != nil {
		d.log.Fatal(err)
	}

	script_path := path.Join(wd, d.config.ScriptsLocation, "downloader")
	storage_location := "."

	process := exec.Command(
		script_path,
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

	reader := bufio.NewReader((stdout))

	var filename string

outerLoop:
	for {
		message, err := reader.ReadBytes('\n')

		d.log.Tracef("Handling script message: %v", string(message))

		var parsedMessage map[string]interface{}

		if err != nil {
			d.log.Errorf("failed to read message from script: %v", err)
			process.Process.Kill()
			return "", errors.New("downloading failed")
		}

		err = json.Unmarshal(message, &parsedMessage)

		if err != nil {
			d.log.Errorf("failed to parse message from script: %v", err)
			process.Process.Kill()
			return "", errors.New("downloading failed")
		}

		t, ok := parsedMessage["type"]

		if !ok {
			d.log.Errorf("message from does not contain \"type\" field")
			process.Process.Kill()
			return "", errors.New("downloading failed")
		}

		switch mtype(t.(float64)) {
		case DownloadingProgress:
			err := d.handleProgressMessage(parsedMessage)
			if err != nil {
				d.log.Errorf("failed to handle progress message: %v, reason: %v",
					parsedMessage, err)
				process.Process.Kill()
				return "", errors.New("downloading failed")
			}
		case DownloadingDone:
			fn, err := d.handleDoneMessage(parsedMessage)

			if err != nil {
				d.log.Errorf("failed to handle progress message: %v, reason: %v",
					parsedMessage, err)
				process.Process.Kill()
				return "", errors.New("downloading failed")
			}

			filename = fn
			break outerLoop
		default:
			d.log.Errorf("no message handler for type: %v", err)
			process.Process.Kill()
			return "", errors.New("downloading failed")
		}
	}

	err = process.Wait()

	if err != nil {
		d.log.Fatal(err)
	}

	d.log.Debugf("Downloading done for file: %v", url)
	return filename, nil
}
