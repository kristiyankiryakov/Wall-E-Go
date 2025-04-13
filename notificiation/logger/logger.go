package logger

import (
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"sync"
)

var (
	instance *logrus.Logger
	once     sync.Once
	logFile  *os.File
)

func NewLogger() *logrus.Logger {
	once.Do(func() {
		instance = logrus.New()

		logPath := filepath.Join(".", "service.log")
		var err error
		logFile, err = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			panic("Failed to open log file: " + err.Error())
		}

		instance.SetOutput(logFile)
		instance.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
		instance.SetLevel(logrus.InfoLevel)
	})

	return instance
}

func Cleanup() {
	if logFile != nil {
		logFile.Close()
	}
}
