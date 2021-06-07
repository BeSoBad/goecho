package tcpserver

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestServer_Base(t *testing.T) {
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
		_, err = fmt.Fprint(c, text)
		require.NoError(t, err)
		message, _ := bufio.NewReader(c).ReadString('\n')
		require.Equal(t, text, message)

		err = server.Shutdown()
		require.NoError(t, err)

		message, _ = bufio.NewReader(c).ReadString('\n')
		require.Equal(t, CloseMessage, message)
	}()
	wg.Wait()
}

func TestServer_MultipleConnections(t *testing.T) {
	connCnt := 1000
	defer goleak.VerifyNone(t)

	server := New(&config, logger)
	err := server.Start()
	require.NoError(t, err)

	acceptWg := &sync.WaitGroup{}
	acceptWg.Add(1)
	go func() {
		defer acceptWg.Done()

		for i := 0; i < connCnt; i++ {
			err = server.Accept(EchoHandler)
			require.NoError(t, err)
		}
	}()

	dialWg := &sync.WaitGroup{}
	dialWg.Add(connCnt)
	for i := 0; i < connCnt; i++ {
		go func(connID int) {
			defer dialWg.Done()

			c, err := net.Dial("tcp", address)
			require.NoError(t, err)

			text := "test_text_1_" + strconv.Itoa(connID) + "\n"
			_, err = fmt.Fprint(c, text)
			require.NoError(t, err)
			message, _ := bufio.NewReader(c).ReadString('\n')
			require.Equal(t, text, message)

			text = "test_text_2_" + strconv.Itoa(connID) + "\n"
			_, err = fmt.Fprint(c, text)
			require.NoError(t, err)
			message, _ = bufio.NewReader(c).ReadString('\n')
			require.Equal(t, text, message)

			message, _ = bufio.NewReader(c).ReadString('\n')
			require.Equal(t, CloseMessage, message)
		}(i)
	}
	acceptWg.Wait()
	err = server.Shutdown()
	require.NoError(t, err)
	dialWg.Wait()
}

func TestServer_StartShutdown(t *testing.T) {
	defer goleak.VerifyNone(t)

	server := New(&config, logger)
	err := server.Accept(EchoHandler)
	require.Equal(t, ErrServerNotStarted, err)

	err = server.Shutdown()
	require.Equal(t, ErrServerNotStarted, err)

	err = server.Start()
	require.NoError(t, err)

	err = server.Start()
	require.Equal(t, ErrServerStarted, err)

	err = server.Shutdown()
	require.NoError(t, err)

	err = server.Shutdown()
	require.Equal(t, ErrServerStopped, err)

	err = server.Start()
	require.Equal(t, ErrServerStopped, err)

	err = server.Accept(EchoHandler)
	require.Equal(t, ErrServerStopped, err)
}
