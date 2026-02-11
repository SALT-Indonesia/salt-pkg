package logmanager

import (
	"io"
	"time"

	"github.com/SALT-Indonesia/salt-pkg/logmanager/internal"
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

func newStandardLogger(debug bool, logDir string, splitLevelOutput bool, masker *internal.JSONMasker) *logrus.Logger {
	l := logrus.New()
	formatter := &logrus.JSONFormatter{}
	l.SetFormatter(formatter)
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
	} else if splitLevelOutput {
		// Discard the default output; the hook handles writing to stdout/stderr
		l.SetOutput(io.Discard)
		l.AddHook(newSplitLevelOutputHook(formatter))
	}

	return l
}
