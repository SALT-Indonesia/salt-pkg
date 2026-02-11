package logmanager

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

// splitLevelOutputHook is a logrus hook that directs log output to different
// writers based on log level, following Twelve-Factor App principles:
// - DEBUG, INFO, TRACE levels → stdout
// - WARN, ERROR, FATAL, PANIC levels → stderr
type splitLevelOutputHook struct {
	stdoutWriter io.Writer
	stderrWriter io.Writer
	formatter    logrus.Formatter
}

func newSplitLevelOutputHook(formatter logrus.Formatter) *splitLevelOutputHook {
	return &splitLevelOutputHook{
		stdoutWriter: os.Stdout,
		stderrWriter: os.Stderr,
		formatter:    formatter,
	}
}

// Levels returns all log levels so the hook fires for every log entry.
func (h *splitLevelOutputHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire writes the log entry to the appropriate writer based on log level.
func (h *splitLevelOutputHook) Fire(entry *logrus.Entry) error {
	formatted, err := h.formatter.Format(entry)
	if err != nil {
		return err
	}

	writer := h.writerForLevel(entry.Level)
	_, err = writer.Write(formatted)
	return err
}

func (h *splitLevelOutputHook) writerForLevel(level logrus.Level) io.Writer {
	switch level {
	case logrus.WarnLevel, logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		return h.stderrWriter
	default:
		return h.stdoutWriter
	}
}
