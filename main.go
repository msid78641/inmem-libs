package main

import (
	"context"
	"fmt"
	"inmem/lib/logger"
	"inmem/shutdown"
	to_do "inmem/src/to-do"
	_ "net/http/pprof"
)

var (
	shutDownChan = make(chan bool)
)

func main() {
	shutdown.AddHook(shutDownHook)
	logger.Dispatch(logger.INFO, "Main app has started running")
	go to_do.TestCacheMixedTraffic5kQPS()
	fmt.Println("Listneing to shutDown channel")
	<-shutDownChan
}

func shutDownHook(ctx context.Context) {
	logger.Dispatch(logger.INFO, "Main app has ended")
	shutDownChan <- true
}
