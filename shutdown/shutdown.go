package shutdown

import (
	"context"
	"fmt"
	"inmem/lib/logger"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Callback func(ctx context.Context)

var (
	mu        sync.Mutex
	callbacks []Callback
	sigChan   = make(chan os.Signal, 1)
)

// default signals
var shutdownSignals = []os.Signal{syscall.SIGINT, syscall.SIGTERM}

func init() {
	signal.Notify(sigChan, shutdownSignals...)
	go listen()
}

// AddHook registers a shutdown callback
func AddHook(cb Callback) {
	mu.Lock()
	defer mu.Unlock()

	callbacks = append(callbacks, cb)
}

func listen() {
	<-sigChan
	execute()
	signal.Stop(sigChan)
}

func execute() {

	// context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// get snapshot of callbacks safely
	mu.Lock()
	cbs := make([]Callback, len(callbacks))
	copy(cbs, callbacks)
	mu.Unlock()

	for _, cb := range cbs {
		safeRun(ctx, cb)
	}

	logger.Dispatch(logger.INFO, "All shutdown hook executed successfully")
}

func safeRun(ctx context.Context, cb Callback) {
	defer func() {
		if r := recover(); r != nil {
			logger.Dispatch(logger.ERROR, fmt.Sprintf("Shutdown hook panic recovered: %v\n", r))

		}
	}()
	cb(ctx)
}
