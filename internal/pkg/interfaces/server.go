package interfaces

type MessageHandler = func(data []byte) ([]byte, error)

type Server interface {
	Accept(handler MessageHandler) error

	Start() error
	Shutdown() error
}
