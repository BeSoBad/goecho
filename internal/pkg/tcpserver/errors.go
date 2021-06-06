package tcpserver

import (
	"errors"
)

var (
	ErrStart            = errors.New("error while starting listening TCP connection")
	ErrShutdown         = errors.New("error while closing TCP connection")
	ErrServerStopped    = errors.New("server is stopped")
	ErrServerNotStarted = errors.New("server is not started")
)
