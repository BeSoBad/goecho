package tcpserver

import (
	"os"
	"strconv"
	"testing"

	log "github.com/sirupsen/logrus"
)

var (
	config  = Config{Host: "0.0.0.0", Port: 8080, BufferSize: 2048}
	address = config.Host + ":" + strconv.Itoa(int(config.Port))
)

var logger *log.Logger

func TestMain(m *testing.M) { // nolint:interfacer
	logger = log.New()

	os.Exit(m.Run())
}
