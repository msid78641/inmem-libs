package logger

import (
	"fmt"
	"time"
)

const (
	ColorReset = "\033[0m"

	ColorRed    = "\033[31m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorGreen  = "\033[32m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
)

func colorForLevel(level Level) string {
	switch level {
	case ERROR:
		return ColorRed
	case WARN:
		return ColorYellow
	case INFO:
		return ColorGreen
	case DEBUG:
		return ColorBlue
	default:
		return ColorWhite
	}
}

func FormatLog(entry *LogEntry) string {
	color := colorForLevel(entry.Level)

	// Convert fields map to "key=value key=value"
	var fields string
	for k, v := range entry.Fields {
		fields += fmt.Sprintf("%s=%s ", k, v)
	}

	return fmt.Sprintf(
		"%s[%s] : [%s] : %s : %s",
		color,
		entry.Time.Format(time.RFC3339),
		entry.Level,
		entry.Msg,
		fields,
		ColorReset,
	)
}

type ConsoleDispatcher struct {
}

func GetConsoleDispatcher() *ConsoleDispatcher {
	return new(ConsoleDispatcher)
}
func (cd *ConsoleDispatcher) Dispatch(l *LogEntry) {
	fmt.Println(FormatLog(l))
}
