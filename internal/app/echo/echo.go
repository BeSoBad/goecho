package echo

import (
	"github.com/BeSoBad/goecho/internal/pkg/interfaces"
	"github.com/BeSoBad/goecho/internal/pkg/tcpserver"
	log "github.com/sirupsen/logrus"
)

type App struct {
	tcpServer interfaces.Server
	logger    *log.Entry
}

const (
	ModuleName        = "echo_app"
	defaultHost       = "0.0.0.0"
	defaultPort       = 8080
	defaultBufferSize = 2048
)

func New(logger *log.Logger) *App {
	config := tcpserver.Config{Host: defaultHost, Port: defaultPort, BufferSize: defaultBufferSize}
	return &App{
		tcpServer: tcpserver.New(&config, logger),
		logger: logger.WithFields(log.Fields{
			"module": ModuleName,
		}),
	}
}

func (a *App) Run() error {
	a.logger.Infof("Echo app is running")
	err := a.tcpServer.Start()
	if err != nil {
		return err
	}
	for {
		err = a.tcpServer.Accept(tcpserver.EchoHandler)
		if err != nil {
			return err
		}
	}
}

func (a *App) Stop() error {
	a.logger.Infof("Echo app is stopping")
	return a.tcpServer.Shutdown()
}
