package tcpserver

import (
	"log"
	"net"
	"strconv"
)

type TCPMessageHandler = func(data []byte) []byte

func EchoHandler(data []byte) []byte {
	return data
}

type IServer interface {
	Start()
	Stop()
}

type Server struct {
}

func (s *Server) Start(handler TCPMessageHandler) {
	host := "0.0.0.0"
	port := 8080
	l, err := net.Listen("tcp", host+":"+strconv.Itoa(port))
	if err != nil {
		log.Panicln(err)
	}
	log.Println("Listening to connections at '"+host+"' on port", strconv.Itoa(port))
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Panicln(err)
		}

		go s.handleRequest(conn, handler)
	}
}

func (s *Server) handleRequest(conn net.Conn, handler TCPMessageHandler) {
	defer conn.Close()
	defer log.Println("Closed connection.")

	log.Println("Accepted new connection.")

	for {
		buf := make([]byte, 1024)
		size, err := conn.Read(buf)
		if err != nil {
			return
		}
		data := handler(buf[:size])
		log.Println("Read new data from connection", data)
		conn.Write(data)
	}
}
