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
	writers []io.Writer
	mutex   *sync.Mutex
}

func (w ThreadsafeWriter) Write(p []byte) (n int, err error) {
	w.mutex.Lock()
	for _, writer := range w.writers {
		n, err = writer.Write(p)
	}
	w.mutex.Unlock()
	return
}

func Init(logDirectory string, logFile string) {
	log := logrus.New()

	level := logrus.TraceLevel
	log.SetLevel(level)

	if _, err := os.Stat(logDirectory); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(logDirectory, os.ModePerm)
		if err != nil {
			log.Fatalf("Error creating directory: %v", err)
		}
	}

	file, err := os.OpenFile(filepath.Join(logDirectory, logFile),
		os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}

	writer = &ThreadsafeWriter{
		writers: []io.Writer{file, os.Stdout},
		mutex:   &sync.Mutex{},
	}

	applicationLogger = logrus.New()
	applicationLogger.SetLevel(level)
	applicationLogger.SetOutput(writer)
	applicationLogger.SetNoLock()
	ApplicationLogger = applicationLogger.WithField("layer", "Application")

	businessLogger = logrus.New()
	businessLogger.SetLevel(level)
	businessLogger.SetOutput(writer)
	businessLogger.SetNoLock()
	BusinessLogger = businessLogger.WithField("layer", "Business")

	presentationLogger = logrus.New()
	presentationLogger.SetLevel(level)
	presentationLogger.SetOutput(writer)
	presentationLogger.SetNoLock()
	PresentationLogger = presentationLogger.WithField("layer", "Presentation")

	dataLogger = logrus.New()
	dataLogger.SetLevel(level)
	dataLogger.SetOutput(writer)
	dataLogger.SetNoLock()
	DataLogger = dataLogger.WithField("layer", "Data")
}

func CloseLogFile() {
	logFile.Close()
}
