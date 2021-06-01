package main

import (
	"github.com/BeSoBad/goecho/internal/tcpserver"
)

func main() {
	server := tcpserver.Server{}
	server.Start(tcpserver.EchoHandler)
}
