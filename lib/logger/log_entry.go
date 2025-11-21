package logger

import (
	"encoding/json"
	"time"
)

type Level string

type LogEntry struct {
	Fields map[string]string
	Msg    string
	Time   time.Time
	Level  Level
}

const (
	INFO  Level = "info"
	WARN  Level = "warn"
	DEBUG Level = "debug"
	ERROR Level = "error"
)

func (l *LogEntry) WithMessage(msg string) *LogEntry {
	l.Msg = msg
	return l
}
func (l *LogEntry) WithField(key string, value string) *LogEntry {
	l.Fields[key] = value
	return l
}

func (l *LogEntry) WithFieldMap(value map[string]string) *LogEntry {
	l.Fields = value
	return l
}
func (l *LogEntry) withTime(time time.Time) *LogEntry {
	l.Time = time
	return l
}

func (l *LogEntry) withLevel(level Level) *LogEntry {
	l.Level = level
	return l
}

func (l *LogEntry) string() string {
	val, _ := json.Marshal(l)
	return string(val)
}

func WithEntry() *LogEntry {
	return &LogEntry{
		Fields: map[string]string{},
	}
}
