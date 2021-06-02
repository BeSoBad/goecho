package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/BeSoBad/goecho/internal/pkg/tcpserver"
	"github.com/sirupsen/logrus"
)

// TODO: SIGTERM, SIGINT handling
func main() {
	config := tcpserver.Config{Host: "0.0.0.0", Port: 8080, BufferSize: 1024}
	logger := logrus.New()
	server := tcpserver.New(&config, logger)
	server.Start()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-ch
		logger.Infoln("Received SIGTERM")
		server.Stop()
	}()

	for {
		server.Accept(tcpserver.EchoHandler)
	}
}
