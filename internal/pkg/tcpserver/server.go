package tcpserver

import (
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/BeSoBad/goecho/internal/pkg/interfaces"
	log "github.com/sirupsen/logrus"
)

const (
	ModuleName     = "tcp_server"
	CloseMessage   = "server closed the connection\n"
	EOFMessage     = "EOF"
	TimeoutMessage = "i/o timeout"

	defaultTimeout = 1 * time.Second
)

type Config struct {
	Host       string
	Port       uint32
	BufferSize uint32
}

type Server struct {
	host string

	listener net.Listener
	logger   *log.Entry

	started chan struct{}
	stopped chan struct{}
	mu      sync.RWMutex
	wg      sync.WaitGroup

	port       uint32
	bufferSize uint32
	connCount  uint32
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
		s.logger.Error("Error starting listening TCP:", err)
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

	s.logger.Infof("Stopping TCP server")
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
		connCh := make(chan net.Conn)

		go func(connCh chan<- net.Conn) {
			conn, err := s.listener.Accept()
			if err != nil {
				s.logger.Error("Error while accepting TCP connection:", err)
				return
			}
			connCh <- conn
		}(connCh)

		select {
		case <-s.stopped:
			return ErrServerStopped
		case conn := <-connCh:
			connID := atomic.AddUint32(&s.connCount, 1)
			s.wg.Add(1)
			s.logger.Infof("New connection accepted [id: %d] [addr: %s]", connID, conn.RemoteAddr())
			go s.startReading(conn, handler, connID)
		}
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
		err := conn.SetReadDeadline(time.Now().Add(defaultTimeout))
		if err != nil {
			s.logger.Errorf("setting deadline error [id: %d]: %s", connID, err)
			return
		}
		size, err := conn.Read(buf)
		if err != nil {
			select {
			case <-s.stopped:
				_, err = conn.Write([]byte(CloseMessage))
				if err != nil {
					s.logger.Errorf("writing close message error [id: %d]: %s", connID, err)
				}
				return
			default:
				switch {
				case err.Error() == EOFMessage:
					s.logger.Infof("Connection dropped by user [id: %d]", connID)
					return
				case strings.HasSuffix(err.Error(), TimeoutMessage):
					continue
				default:
					s.logger.Errorf("Reading error: %s", err.Error())
					return
				}
			}
		}
		data := handler(buf[:size])
		s.logger.Infof("Received data [id: %d]: %s", connID, string(data))
		_, err = conn.Write(data)
		if err != nil {
			s.logger.Errorf("writing message error [id: %d] [data: %s]: %s", connID, data, err)
		}
	}
}
