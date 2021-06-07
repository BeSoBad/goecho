package tcpserver

import (
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/BeSoBad/goecho/internal/pkg/interfaces"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

const (
	ModuleName              = "tcp_server"
	CloseMessage            = "server closed the connection\n"
	EOFMessage              = "EOF"
	TimeoutMessage          = "i/o timeout"
	ClosedConnectionMessage = "use of closed network connection"

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

	started  chan struct{}
	stopped  chan struct{}
	mu       sync.RWMutex
	readWg   sync.WaitGroup
	acceptWg sync.WaitGroup

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

	select {
	case <-s.stopped:
		return ErrServerStopped
	default:
	}

	select {
	case <-s.started:
		return ErrServerStarted
	default:
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
}

func (s *Server) Shutdown() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	select {
	case <-s.stopped:
		return ErrServerStopped
	default:
	}

	select {
	case <-s.started:
		s.logger.Infof("Stopping TCP server")
		close(s.stopped)
		s.readWg.Wait()
		err := s.listener.Close()
		s.acceptWg.Wait()
		if err != nil {
			s.logger.Error("Error while closing TCP listener:", err)
			return ErrShutdown
		}

		return nil
	default:
		return ErrServerNotStarted
	}
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

		s.acceptWg.Add(1)
		go func(connCh chan<- net.Conn) {
			defer s.acceptWg.Done()

			conn, err := s.listener.Accept()
			if err != nil {
				if !strings.HasSuffix(err.Error(), ClosedConnectionMessage) {
					s.logger.Error("Error while accepting TCP connection:", err)
				}
				return
			}
			connCh <- conn
		}(connCh)

		select {
		case <-s.stopped:
			return ErrServerStopped
		case conn := <-connCh:
			atomic.AddUint32(&s.connCount, 1)
			connID, _ := uuid.NewUUID()
			s.readWg.Add(1)
			s.logger.Infof("New connection accepted [id: %s] [addr: %s]", connID.String(), conn.RemoteAddr())
			go s.startReading(conn, handler, connID.String())
		}
	default:
		return ErrServerNotStarted
	}

	return nil
}

func (s *Server) startReading(conn net.Conn, handler interfaces.MessageHandler, connID string) {
	defer s.readWg.Done()
	defer atomic.StoreUint32(&s.connCount, atomic.LoadUint32(&s.connCount)-1)
	defer func() {
		s.logger.Infof("Closing connection [id: %s]", connID)
		_, err := conn.Write([]byte(CloseMessage))
		if err != nil {
			s.logger.Errorf("Writing close message error [id: %s]: %s", connID, err)
		}
		err = conn.Close()
		if err != nil {
			s.logger.Errorf("Closing connection error [id: %s]: %s", connID, err)
		}
	}()

	for {
		buf := make([]byte, s.bufferSize)
		err := conn.SetReadDeadline(time.Now().Add(defaultTimeout))
		if err != nil {
			s.logger.Errorf("Setting deadline error [id: %s]: %s", connID, err)
			return
		}
		size, err := conn.Read(buf)
		if err != nil {
			select {
			case <-s.stopped:
				return
			default:
				switch {
				case err.Error() == EOFMessage:
					s.logger.Infof("Connection terminated by user [id: %s]", connID)
					return
				case strings.HasSuffix(err.Error(), TimeoutMessage):
					continue
				default:
					s.logger.Errorf("Reading error: %s", err.Error())
					return
				}
			}
		}
		sizeBuf := buf[:size]
		s.logger.Infof("Received data [id: %s]: %s", connID, string(sizeBuf))
		handledBuf, err := handler(sizeBuf)
		if err != nil {
			s.logger.Errorf("Handling message error [id: %s] [data: %s]: %s", connID, handledBuf, err)
			return
		}
		size, err = conn.Write(handledBuf)
		if err != nil {
			s.logger.Errorf("Writing message error [id: %s] [data: %s]: %s", connID, handledBuf, err)
			return
		}
		s.logger.Infof("Sent data [id: %s]: %s", connID, string(handledBuf[:size]))
	}
}
