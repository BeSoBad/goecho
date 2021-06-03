package echo

import (
	"github.com/BeSoBad/goecho/internal/pkg/interfaces"
	"github.com/BeSoBad/goecho/internal/pkg/tcpserver"
	log "github.com/sirupsen/logrus"
)

type EchoApp struct {
	tcpServer interfaces.Server
	logger    *log.Entry
}

const (
	ModuleName        = "echo_app"
	defaultHost       = "0.0.0.0"
	defaultPort       = 8080
	defaultBufferSize = 2048
)

func New(logger *log.Logger) *EchoApp {
	config := tcpserver.Config{Host: defaultHost, Port: defaultPort, BufferSize: defaultBufferSize}
	return &EchoApp{
		tcpServer: tcpserver.New(&config, logger),
		logger: logger.WithFields(log.Fields{
			"module": ModuleName,
		}),
	}
}

func (e *EchoApp) Run() error {
	e.logger.Infof("Echo app is running")
	err := e.tcpServer.Start()
	if err != nil {
		return err
	}
	for {
		err = e.tcpServer.Accept(EchoHandler)
		if err != nil {
			return err
		}
	}
}

func (e *EchoApp) Stop() error {
	e.logger.Infof("Echo app is stopping")
	return e.tcpServer.Shutdown()
}
