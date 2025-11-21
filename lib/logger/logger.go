package logger

import (
	"time"
)

type LogEntryType interface {
	string | *LogEntry
}

type LogDispatcher interface {
	Dispatch(l *LogEntry)
}

type Logger struct {
	logDispatcher LogDispatcher
}

var l = &Logger{
	logDispatcher: GetDispatcher(),
}

func getLogEntry[T LogEntryType](le T) *LogEntry {
	var logEntry *LogEntry = nil
	switch x := any(le).(type) {
	case string:
		logEntry = WithEntry().WithMessage(x)
	case *LogEntry:
		logEntry = x
	}
	return logEntry
}

func Dispatch[T LogEntryType](logLevel Level, le T) {
	logEntry := getLogEntry(le).
		withTime(time.Now()).
		withLevel(logLevel)
	l.logDispatcher.Dispatch(logEntry)
}
