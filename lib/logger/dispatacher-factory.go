package logger

func GetDispatcher() LogDispatcher {
	// Should be changed to switch case
	//if _, ok := config["fileDispatcher"]; ok {
	//	return GetFileDispatcher("asfasf")
	//}
	return GetFileDispatcher("logs/app.log")
}
