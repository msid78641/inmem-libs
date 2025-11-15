package logger

import (
	logger "inmem/lib/logger/internal"
	"time"
)

type LogDispatcher interface {
	Dispatch(l *LogEntry)
}

type Logger struct {
	logDispatcher LogDispatcher
}

var l = &Logger{
	logDispatcher: logger.GetDispatcher(),
}

func Dispatch(logLevel Level, logEntry *LogEntry) {
	l.logDispatcher.Dispatch(logEntry.withTime(time.Now()).withLevel(logLevel))
}
