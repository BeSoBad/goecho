package tcpserver

import (
	"errors"
)

var (
	ErrStart  = errors.New("Error while starting listening TCP connection")
	ErrAccept = errors.New("Accept error")
	ErrClose  = errors.New("Err close")
)
