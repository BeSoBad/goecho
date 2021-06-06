package tcpserver

import (
	"bufio"
	"fmt"
	"net"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestServer_BaseScenario(t *testing.T) {
	defer goleak.VerifyNone(t)

	server := New(&config, logger)
	err := server.Start()
	require.NoError(t, err)

	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()

		err = server.Accept(EchoHandler)
		require.NoError(t, err)
	}()

	go func() {
		defer wg.Done()

		c, err := net.Dial("tcp", address)
		require.NoError(t, err)

		text := "test_text\n"
		fmt.Fprint(c, text)
		message, _ := bufio.NewReader(c).ReadString('\n')
		require.Equal(t, text, message)

		err = server.Shutdown()
		require.NoError(t, err)

		message, _ = bufio.NewReader(c).ReadString('\n')
		require.Equal(t, CloseMessage, message)
	}()
	wg.Wait()
}

func TestServer_StartShutdown(t *testing.T) {
	defer goleak.VerifyNone(t)

	server := New(&config, logger)
	err := server.Accept(EchoHandler)
	require.Equal(t, ErrServerNotStarted, err)
	err = server.Start()
	require.NoError(t, err)
	err = server.Shutdown()
	require.NoError(t, err)
	err = server.Accept(EchoHandler)
	require.Equal(t, ErrServerStopped, err)
}
