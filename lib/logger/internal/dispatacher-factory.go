package logger

import "inmem/lib/logger"

func GetDispatcher() logger.LogDispatcher {
	// Should be changed to switch case
	//if _, ok := config["fileDispatcher"]; ok {
	//	return GetFileDispatcher("asfasf")
	//}
	return GetConsoleDispatcher()
}
