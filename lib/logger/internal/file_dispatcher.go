package logger

import "inmem/lib/logger"

type FileDispatcher struct {
	filePath string
}

func GetFileDispatcher(filePath string) *FileDispatcher {
	return &FileDispatcher{
		filePath: filePath,
	}
}

func (fd *FileDispatcher) Dispatch(l *logger.LogEntry) {

}
