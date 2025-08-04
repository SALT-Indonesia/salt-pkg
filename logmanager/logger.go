package logmanager

import (
	"github.com/SALT-Indonesia/salt-pkg/logmanager/internal"
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
	"time"
)

func newStandardLogger(debug bool, logDir string, masker *internal.JSONMasker) *logrus.Logger {
	l := logrus.New()
	l.SetFormatter(&logrus.JSONFormatter{})
	l.SetLevel(logrus.InfoLevel)
	l.AddHook(masker.LogrusMiddleware())
	if debug {
		l.SetLevel(logrus.DebugLevel)
	}

	if logDir != "" {
		logFilename := time.Now().Format("2006-01-02") + ".log"
		logFilePath := logDir + "/" + logFilename

		l.SetOutput(&lumberjack.Logger{
			Filename: logFilePath,
		})
	}

	return l
}
