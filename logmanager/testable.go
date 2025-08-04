package logmanager

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

// TestableApplication extends Application with testing capabilities
type TestableApplication struct {
	*Application
	testHook *TestHook
}

// NewTestableApplication creates a new Application instance with testing capabilities.
// This allows tests to capture and inspect logged entries.
func NewTestableApplication(opts ...Option) *TestableApplication {
	app := NewApplication(opts...)
	testHook := NewTestHook()

	// Add the test hook to the logger
	app.logger.AddHook(testHook)

	return &TestableApplication{
		Application: app,
		testHook:    testHook,
	}
}

// GetLoggedEntries returns all logged entries captured by the test hook
func (ta *TestableApplication) GetLoggedEntries() []*logrus.Entry {
	if ta.testHook == nil {
		return nil
	}
	return ta.testHook.AllEntries()
}

// GetLastLoggedEntry returns the last logged entry
func (ta *TestableApplication) GetLastLoggedEntry() *logrus.Entry {
	if ta.testHook == nil {
		return nil
	}
	return ta.testHook.LastEntry()
}

// ResetLoggedEntries clears all captured log entries
func (ta *TestableApplication) ResetLoggedEntries() {
	if ta.testHook != nil {
		ta.testHook.Reset()
	}
}

// GetLoggedField returns the value of a specific field from the last logged entry
func (ta *TestableApplication) GetLoggedField(fieldName string) interface{} {
	entry := ta.GetLastLoggedEntry()
	if entry == nil {
		return nil
	}
	return entry.Data[fieldName]
}

// GetLoggedFields returns all fields from the last logged entry
func (ta *TestableApplication) GetLoggedFields() logrus.Fields {
	entry := ta.GetLastLoggedEntry()
	if entry == nil {
		return nil
	}
	return entry.Data
}

// HasLoggedField checks if a specific field exists in the last logged entry
func (ta *TestableApplication) HasLoggedField(fieldName string) bool {
	entry := ta.GetLastLoggedEntry()
	if entry == nil {
		return false
	}
	_, exists := entry.Data[fieldName]
	return exists
}

// GetLoggedLevel returns the log level of the last logged entry
func (ta *TestableApplication) GetLoggedLevel() logrus.Level {
	entry := ta.GetLastLoggedEntry()
	if entry == nil {
		return logrus.PanicLevel // Return an invalid level to indicate no entry
	}
	return entry.Level
}

// GetLoggedMessage returns the message of the last logged entry
func (ta *TestableApplication) GetLoggedMessage() string {
	entry := ta.GetLastLoggedEntry()
	if entry == nil {
		return ""
	}
	return entry.Message
}

// CountLoggedEntries returns the number of logged entries
func (ta *TestableApplication) CountLoggedEntries() int {
	return len(ta.GetLoggedEntries())
}

// GetLoggedEntriesWithLevel returns all logged entries with a specific level
func (ta *TestableApplication) GetLoggedEntriesWithLevel(level logrus.Level) []*logrus.Entry {
	entries := ta.GetLoggedEntries()
	var filtered []*logrus.Entry
	for _, entry := range entries {
		if entry.Level == level {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

// GetLoggedEntriesWithField returns all logged entries that contain a specific field
func (ta *TestableApplication) GetLoggedEntriesWithField(fieldName string) []*logrus.Entry {
	entries := ta.GetLoggedEntries()
	var filtered []*logrus.Entry
	for _, entry := range entries {
		if _, exists := entry.Data[fieldName]; exists {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

// AddTestHook adds a test hook to an existing Application for testing purposes.
// This is useful when you want to make an existing Application testable.
func (app *Application) AddTestHook() *TestHook {
	if app == nil || app.logger == nil {
		return nil
	}

	testHook := NewTestHook()
	app.logger.AddHook(testHook)
	return testHook
}
