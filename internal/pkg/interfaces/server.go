package interfaces

type MessageHandler = func(data []byte) []byte

type Server interface {
	Accept(handler MessageHandler) error

	Start() error
	Shutdown() error
}
