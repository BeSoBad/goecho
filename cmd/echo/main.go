package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/BeSoBad/goecho/internal/app/echo"
	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()
	app := echo.New(logger)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	signal.Notify(ch, os.Interrupt, syscall.SIGINT)
	go func() {
		<-ch
		logger.Infoln("Received exiting signal")
		app.Stop()
	}()

	err := app.Run()
	if err != nil {
		logger.Infof("Error while running echo app: %s", err)
	}
}
