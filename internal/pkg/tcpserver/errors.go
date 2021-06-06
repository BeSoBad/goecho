package tcpserver

import (
	"errors"
)

var (
	ErrStart            = errors.New("error while starting listening TCP connection")
	ErrClose            = errors.New("err close")
	ErrServerStopped    = errors.New("server is stopped")
	ErrServerNotStarted = errors.New("server is not started")
)
