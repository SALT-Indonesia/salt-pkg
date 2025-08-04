package testdata

import (
	"github.com/sirupsen/logrus"
	"sync"
)

// TestHook is a hook designed for testing logrus logging
type TestHook struct {
	Entries []*logrus.Entry
	mu      sync.Mutex
}

// NewTestHook creates a new test hook for logrus
func NewTestHook() *TestHook {
	return &TestHook{
		Entries: make([]*logrus.Entry, 0),
	}
}

// Fire implements logrus.Hook.Fire
func (hook *TestHook) Fire(entry *logrus.Entry) error {
	hook.mu.Lock()
	defer hook.mu.Unlock()

	// Make a copy of the entry to avoid race conditions
	newEntry := &logrus.Entry{
		Logger:  entry.Logger,
		Data:    make(logrus.Fields),
		Time:    entry.Time,
		Level:   entry.Level,
		Message: entry.Message,
	}

	for k, v := range entry.Data {
		newEntry.Data[k] = v
	}

	hook.Entries = append(hook.Entries, newEntry)
	return nil
}

// Levels implements logrus.Hook.Levels
func (hook *TestHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Reset clears all entries
func (hook *TestHook) Reset() {
	hook.mu.Lock()
	defer hook.mu.Unlock()
	hook.Entries = make([]*logrus.Entry, 0)
}

// LastEntry returns the last entry that was logged
func (hook *TestHook) LastEntry() *logrus.Entry {
	hook.mu.Lock()
	defer hook.mu.Unlock()

	if len(hook.Entries) == 0 {
		return nil
	}

	return hook.Entries[len(hook.Entries)-1]
}

// AllEntries returns all entries
func (hook *TestHook) AllEntries() []*logrus.Entry {
	hook.mu.Lock()
	defer hook.mu.Unlock()

	return hook.Entries
}
