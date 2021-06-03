package tcpserver

import (
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/BeSoBad/goecho/internal/pkg/interfaces"
	log "github.com/sirupsen/logrus"
)

const (
	ModuleName     = "tcp_server"
	CloseMessage   = "\nserver closed the connection\n"
	defaultTimeout = 1 * time.Second
)

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
	wg      sync.WaitGroup
	started chan struct{}
	stopped chan struct{}

	connCount uint32
}

func New(config *Config, logger *log.Logger) *Server {
	return &Server{
		host:       config.Host,
		port:       config.Port,
		bufferSize: config.BufferSize,
		started:    make(chan struct{}),
		stopped:    make(chan struct{}),

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

	s.logger.Infof("TCP server started listening connections [host: %s] [port: %s]", s.host, strconv.Itoa(int(s.port)))
	close(s.started)
	return nil
}

func (s *Server) Shutdown() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	close(s.stopped)
	s.wg.Wait()
	err := s.listener.Close()
	if err != nil {
		s.logger.Error("Error while closing TCP listener:", err)
		return ErrClose
	}

	return nil
}

func (s *Server) Accept(handler interfaces.MessageHandler) error {
	select {
	case <-s.stopped:
		return ErrServerStopped
	default:
	}

	select {
	case <-s.started:
		conn, err := s.listener.Accept()
		if err != nil {
			s.logger.Error("Error while accepting TCP connection:", err)
			return ErrAccept
		}

		connID := atomic.AddUint32(&s.connCount, 1)
		s.wg.Add(1)
		s.logger.Infof("New connection is accepted [id: %d] [addr: %s]", connID, conn.RemoteAddr())
		go s.startReading(conn, handler, connID)
	default:
		return ErrServerNotStarted
	}

	return nil
}

func (s *Server) startReading(conn net.Conn, handler interfaces.MessageHandler, connID uint32) {
	defer s.wg.Done()
	defer conn.Close()
	defer atomic.StoreUint32(&s.connCount, atomic.LoadUint32(&s.connCount)-1)
	defer s.logger.Infof("Closing connection [id: %d]", connID)

	for {
		buf := make([]byte, s.bufferSize)
		conn.SetReadDeadline(time.Now().Add(defaultTimeout))
		size, err := conn.Read(buf)
		if err != nil {
			select {
			case <-s.stopped:
				conn.Write([]byte(CloseMessage))
				return
			default:
				continue
			}
		}
		data := handler(buf[:size])
		s.logger.Infof("Received data [id: %d]: %s", connID, string(data))
		conn.Write(data)
	}
}
