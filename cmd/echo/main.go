package main

import (
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/BeSoBad/goecho/internal/app/echo"
	"github.com/BeSoBad/goecho/internal/pkg/tcpserver"
	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()
	app := echo.New(logger)

	wg := sync.WaitGroup{}
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	wg.Add(1)
	var exitStatus uint32

	go func() {
		defer wg.Done()

		<-ch
		logger.Infoln("Received exiting signal")
		err := app.Stop()
		if err != nil {
			logger.Errorf("Error stopping echo app: %s", err)
			atomic.StoreUint32(&exitStatus, 1)
		}
	}()

	err := app.Run()
	if err != nil && err != tcpserver.ErrServerStopped {
		logger.Errorf("Error while running echo app: %s", err)
		atomic.StoreUint32(&exitStatus, 1)
	}
	wg.Wait()

	os.Exit(int(exitStatus))
}
