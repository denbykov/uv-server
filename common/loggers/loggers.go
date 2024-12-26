package loggers

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/sirupsen/logrus"
)

var writer *ThreadsafeWriter
var logFile *os.File

var applicationLogger *logrus.Logger
var ApplicationLogger *logrus.Entry

var businessLogger *logrus.Logger
var BusinessLogger *logrus.Entry

var presentationLogger *logrus.Logger
var PresentationLogger *logrus.Entry

var dataLogger *logrus.Logger
var DataLogger *logrus.Entry

type ThreadsafeWriter struct {
	Writer io.Writer
	Mutex  *sync.Mutex
}

func (w ThreadsafeWriter) Write(p []byte) (n int, err error) {
	w.Mutex.Lock()
	n, err = w.Writer.Write(p)
	w.Mutex.Unlock()
	return
}

func Init() {
	logDirectory := "logs"

	log := logrus.New()

	if _, err := os.Stat(logDirectory); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(logDirectory, os.ModePerm)
		if err != nil {
			log.Fatalf("Error creating directory: %v", err)
		}
	}

	logFile, err := os.OpenFile(filepath.Join(logDirectory, "log.txt"),
		os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}

	writer = &ThreadsafeWriter{
		Writer: logFile,
		Mutex:  &sync.Mutex{},
	}

	applicationLogger = logrus.New()
	applicationLogger.SetOutput(writer)
	applicationLogger.SetNoLock()
	ApplicationLogger = applicationLogger.WithField("name", "Application")

	businessLogger = logrus.New()
	businessLogger.SetOutput(writer)
	businessLogger.SetNoLock()
	BusinessLogger = businessLogger.WithField("name", "Business")

	presentationLogger = logrus.New()
	presentationLogger.SetOutput(writer)
	presentationLogger.SetNoLock()
	PresentationLogger = presentationLogger.WithField("name", "Presentation")

	dataLogger = logrus.New()
	dataLogger.SetOutput(writer)
	dataLogger.SetNoLock()
	DataLogger = dataLogger.WithField("name", "Data")
}

func CloseLogFile() {
	logFile.Close()
}
