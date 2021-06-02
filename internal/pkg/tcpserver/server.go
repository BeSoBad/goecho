package tcpserver

import (
	"net"
	"strconv"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	ModuleName = "tcp_server"
)

type TCPMessageHandler = func(data []byte) []byte

func EchoHandler(data []byte) []byte {
	return data
}

type IServer interface {
	Start()
	Stop()
}

type Config struct {
	Host       string
	Port       uint32
	BufferSize uint32
}

type Server struct {
	host       string
	port       uint32
	bufferSize uint32

	logger   *log.Entry
	listener net.Listener

	mu      sync.RWMutex
	started chan struct{}
	closed  chan struct{}
}

func New(config *Config, logger *log.Logger) *Server {
	return &Server{host: config.Host, port: config.Port, bufferSize: config.BufferSize,
		logger: logger.WithFields(log.Fields{
			"module": ModuleName,
		})}
}

func (s *Server) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	listener, err := net.Listen("tcp", s.host+":"+strconv.Itoa(int(s.port)))
	if err != nil {
		s.logger.Error("Error while starting listening TCP:", err)
		return ErrStart
	}
	s.listener = listener

	s.logger.Info("Listening to connections at '"+s.host+"' on port ", strconv.Itoa(int(s.port)))
	return nil
}

func (s *Server) Accept(handler TCPMessageHandler) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	conn, err := s.listener.Accept()
	if err != nil {
		s.logger.Error("Error while accepting:", err)
		return ErrAccept
	}

	s.logger.Info("New connection accepted: ", conn.RemoteAddr())
	go s.startReading(conn, handler)

	return nil
}

func (s *Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.listener.Close()
	if err != nil {
		s.logger.Error("Error while closing server:", err)
		return ErrClose
	}

	close(s.closed)
	return nil
}

func (s *Server) startReading(conn net.Conn, handler TCPMessageHandler) {
	defer conn.Close()
	defer s.logger.Println("Closed connection.")

	for {
		buf := make([]byte, s.bufferSize)
		conn.SetReadDeadline(time.Now().Add(time.Second * 2))
		size, err := conn.Read(buf)
		if err != nil {
			select {
			case <-s.closed:
				conn.Write([]byte("server closed the connection"))
				return
			default:
				continue
			}
		}
		data := handler(buf[:size])
		s.logger.Println("Read new data from connection", data)
		conn.Write(data)
	}
}
