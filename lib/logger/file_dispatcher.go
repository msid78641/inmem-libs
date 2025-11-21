package logger

import (
	"fmt"
	"log"
	"os"
)

type FileDispatcher struct {
	logger   *log.Logger
	filePath string
}

func GetFileDispatcher(filePath string) *FileDispatcher {
	// Create logs folder if not exists
	os.MkdirAll("logs", 0755)

	// Create/open log file
	f, _ := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	return &FileDispatcher{
		filePath: filePath,
		logger:   log.New(f, "", log.LstdFlags),
	}
}

func (fd *FileDispatcher) Dispatch(l *LogEntry) {
	fd.logger.Println(l.string())
	fmt.Println("Printing to the file dispatcher ", l.string())
}
