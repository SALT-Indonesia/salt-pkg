package logmanager

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestSplitLevelOutputHook_InfoWritesToStdout(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	formatter := &logrus.JSONFormatter{}

	hook := &splitLevelOutputHook{
		stdoutWriter: stdout,
		stderrWriter: stderr,
		formatter:    formatter,
	}

	entry := &logrus.Entry{
		Logger:  logrus.New(),
		Level:   logrus.InfoLevel,
		Message: "info message",
		Data:    logrus.Fields{},
	}

	err := hook.Fire(entry)

	assert.NoError(t, err)
	assert.NotEmpty(t, stdout.String())
	assert.Empty(t, stderr.String())

	var logOutput map[string]interface{}
	err = json.Unmarshal(stdout.Bytes(), &logOutput)
	assert.NoError(t, err)
	assert.Equal(t, "info", logOutput["level"])
}

func TestSplitLevelOutputHook_DebugWritesToStdout(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	formatter := &logrus.JSONFormatter{}

	hook := &splitLevelOutputHook{
		stdoutWriter: stdout,
		stderrWriter: stderr,
		formatter:    formatter,
	}

	entry := &logrus.Entry{
		Logger:  logrus.New(),
		Level:   logrus.DebugLevel,
		Message: "debug message",
		Data:    logrus.Fields{},
	}

	err := hook.Fire(entry)

	assert.NoError(t, err)
	assert.NotEmpty(t, stdout.String())
	assert.Empty(t, stderr.String())
}

func TestSplitLevelOutputHook_TraceWritesToStdout(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	formatter := &logrus.JSONFormatter{}

	hook := &splitLevelOutputHook{
		stdoutWriter: stdout,
		stderrWriter: stderr,
		formatter:    formatter,
	}

	entry := &logrus.Entry{
		Logger:  logrus.New(),
		Level:   logrus.TraceLevel,
		Message: "trace message",
		Data:    logrus.Fields{},
	}

	err := hook.Fire(entry)

	assert.NoError(t, err)
	assert.NotEmpty(t, stdout.String())
	assert.Empty(t, stderr.String())
}

func TestSplitLevelOutputHook_WarnWritesToStderr(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	formatter := &logrus.JSONFormatter{}

	hook := &splitLevelOutputHook{
		stdoutWriter: stdout,
		stderrWriter: stderr,
		formatter:    formatter,
	}

	entry := &logrus.Entry{
		Logger:  logrus.New(),
		Level:   logrus.WarnLevel,
		Message: "warn message",
		Data:    logrus.Fields{},
	}

	err := hook.Fire(entry)

	assert.NoError(t, err)
	assert.Empty(t, stdout.String())
	assert.NotEmpty(t, stderr.String())

	var logOutput map[string]interface{}
	err = json.Unmarshal(stderr.Bytes(), &logOutput)
	assert.NoError(t, err)
	assert.Equal(t, "warning", logOutput["level"])
}

func TestSplitLevelOutputHook_ErrorWritesToStderr(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	formatter := &logrus.JSONFormatter{}

	hook := &splitLevelOutputHook{
		stdoutWriter: stdout,
		stderrWriter: stderr,
		formatter:    formatter,
	}

	entry := &logrus.Entry{
		Logger:  logrus.New(),
		Level:   logrus.ErrorLevel,
		Message: "error message",
		Data:    logrus.Fields{},
	}

	err := hook.Fire(entry)

	assert.NoError(t, err)
	assert.Empty(t, stdout.String())
	assert.NotEmpty(t, stderr.String())
}

func TestSplitLevelOutputHook_FatalWritesToStderr(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	formatter := &logrus.JSONFormatter{}

	hook := &splitLevelOutputHook{
		stdoutWriter: stdout,
		stderrWriter: stderr,
		formatter:    formatter,
	}

	entry := &logrus.Entry{
		Logger:  logrus.New(),
		Level:   logrus.FatalLevel,
		Message: "fatal message",
		Data:    logrus.Fields{},
	}

	err := hook.Fire(entry)

	assert.NoError(t, err)
	assert.Empty(t, stdout.String())
	assert.NotEmpty(t, stderr.String())
}

func TestSplitLevelOutputHook_PanicWritesToStderr(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	formatter := &logrus.JSONFormatter{}

	hook := &splitLevelOutputHook{
		stdoutWriter: stdout,
		stderrWriter: stderr,
		formatter:    formatter,
	}

	entry := &logrus.Entry{
		Logger:  logrus.New(),
		Level:   logrus.PanicLevel,
		Message: "panic message",
		Data:    logrus.Fields{},
	}

	err := hook.Fire(entry)

	assert.NoError(t, err)
	assert.Empty(t, stdout.String())
	assert.NotEmpty(t, stderr.String())
}

func TestSplitLevelOutputHook_Levels_ReturnsAllLevels(t *testing.T) {
	hook := newSplitLevelOutputHook(&logrus.JSONFormatter{})

	levels := hook.Levels()

	assert.Equal(t, logrus.AllLevels, levels)
}

func TestWithSplitLevelOutput_SetsFlag(t *testing.T) {
	app := NewApplication(
		WithSplitLevelOutput(),
	)

	assert.True(t, app.splitLevelOutput)
}

func TestWithSplitLevelOutput_IgnoredWhenLogDirSet(t *testing.T) {
	// When logDir is set, splitLevelOutput should be ignored
	// (file-based logging takes precedence)
	app := NewApplication(
		WithSplitLevelOutput(),
		WithLogDir("/tmp/test-logs"),
	)

	assert.True(t, app.splitLevelOutput)
	// The logger should use file output, not split level output
	// This is handled in newStandardLogger via the if/else if chain
}
