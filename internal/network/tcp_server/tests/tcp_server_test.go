package tests

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	appconfig "github.com/WithSoull/in-memory-database/internal/config/app"
	derrors "github.com/WithSoull/in-memory-database/internal/domainerrors"
	tcpserver "github.com/WithSoull/in-memory-database/internal/network/tcp_server"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// echoHandler returns the request bytes unchanged.
func echoHandler(_ context.Context, req []byte) []byte {
	return req
}

// freeAddr finds a free TCP address by briefly listening on it.
func freeAddr(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := l.Addr().String()
	require.NoError(t, l.Close())
	return addr
}

// startServer creates a server with the given config and handler, starts it in the
// background, and cancels it via t.Cleanup. Returns the address to connect to.
func startServer(t *testing.T, cfg appconfig.NetworkConfig, handler func(context.Context, []byte) []byte) string {
	t.Helper()
	cfg.Address = freeAddr(t)

	srv, err := tcpserver.NewServer(cfg, zap.NewNop())
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	go srv.HandleQueries(ctx, handler)

	return cfg.Address
}

// defaultCfg returns a NetworkConfig suitable for tests.
func defaultCfg() appconfig.NetworkConfig {
	return appconfig.NetworkConfig{
		MaxConnections:      100,
		MaxMessageSizeBytes: 4 * 1024,
		IdleTimeoutDuration: 5 * time.Minute,
	}
}

// ---- Constructor tests ----

func TestNewServer_NilLogger(t *testing.T) {
	t.Parallel()

	cfg := defaultCfg()
	cfg.Address = freeAddr(t)

	srv, err := tcpserver.NewServer(cfg, nil)
	require.ErrorIs(t, err, derrors.ErrInvalidLogger)
	require.Nil(t, srv)
}

func TestNewServer_InvalidAddress(t *testing.T) {
	t.Parallel()

	cfg := defaultCfg()
	cfg.Address = "not-valid:99999"

	srv, err := tcpserver.NewServer(cfg, zap.NewNop())
	require.Error(t, err)
	require.Nil(t, srv)
}

// ---- HandleQueries tests ----

func TestHandleQueries_EchoSingleRequest(t *testing.T) {
	t.Parallel()

	addr := startServer(t, defaultCfg(), echoHandler)

	conn, err := net.Dial("tcp", addr)
	require.NoError(t, err)
	defer conn.Close()

	_, err = fmt.Fprintln(conn, "hello")
	require.NoError(t, err)

	line, err := bufio.NewReader(conn).ReadString('\n')
	require.NoError(t, err)
	require.Equal(t, "hello\n", line)
}

func TestHandleQueries_MultipleRequests(t *testing.T) {
	t.Parallel()

	addr := startServer(t, defaultCfg(), echoHandler)

	conn, err := net.Dial("tcp", addr)
	require.NoError(t, err)
	defer conn.Close()

	reader := bufio.NewReader(conn)

	for i := 0; i < 5; i++ {
		msg := fmt.Sprintf("msg-%d", i)
		_, err := fmt.Fprintln(conn, msg)
		require.NoError(t, err)

		line, err := reader.ReadString('\n')
		require.NoError(t, err)
		require.Equal(t, msg+"\n", line)
	}
}

func TestHandleQueries_GracefulShutdown(t *testing.T) {
	t.Parallel()

	cfg := defaultCfg()
	cfg.Address = freeAddr(t)

	srv, err := tcpserver.NewServer(cfg, zap.NewNop())
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		srv.HandleQueries(ctx, echoHandler)
		close(done)
	}()

	cancel()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("HandleQueries did not return after context cancellation")
	}

	// After shutdown the listener is closed; new connections must fail.
	_, err = net.DialTimeout("tcp", cfg.Address, 500*time.Millisecond)
	require.Error(t, err)
}

func TestHandleQueries_MaxConnections(t *testing.T) {
	t.Parallel()

	cfg := defaultCfg()
	cfg.MaxConnections = 1
	cfg.IdleTimeoutDuration = 30 * time.Second

	addr := startServer(t, cfg, echoHandler)

	// First connection: acquire the only semaphore slot.
	conn1, err := net.Dial("tcp", addr)
	require.NoError(t, err)
	defer conn1.Close()

	// Exchange a request/response to make sure the server has fully accepted conn1
	// (semaphore slot taken, handleConnection goroutine running and waiting for more input).
	_, err = fmt.Fprintln(conn1, "hello")
	require.NoError(t, err)
	line, err := bufio.NewReader(conn1).ReadString('\n')
	require.NoError(t, err)
	require.Equal(t, "hello\n", line)

	// Second connection: semaphore is full — server must reject it immediately.
	conn2, err := net.Dial("tcp", addr)
	require.NoError(t, err)
	defer conn2.Close()

	conn2.SetDeadline(time.Now().Add(2 * time.Second))
	buf := make([]byte, 1)
	_, err = conn2.Read(buf)
	require.Error(t, err, "second connection should be rejected (EOF or reset)")
}

func TestHandleQueries_MessageTooLarge(t *testing.T) {
	t.Parallel()

	cfg := defaultCfg()
	cfg.MaxMessageSizeBytes = 10 // intentionally tiny

	addr := startServer(t, cfg, echoHandler)

	conn, err := net.Dial("tcp", addr)
	require.NoError(t, err)
	defer conn.Close()

	// Send a message longer than MaxMessageSizeBytes.
	_, err = fmt.Fprint(conn, strings.Repeat("x", 20)+"\n")
	require.NoError(t, err)

	line, err := bufio.NewReader(conn).ReadString('\n')
	require.NoError(t, err)
	require.Equal(t, "[error] message too large\n", line)
}

func TestHandleQueries_MessageTooLarge_ConnectionRemainsUsable(t *testing.T) {
	t.Parallel()

	cfg := defaultCfg()
	cfg.MaxMessageSizeBytes = 10

	addr := startServer(t, cfg, echoHandler)

	conn, err := net.Dial("tcp", addr)
	require.NoError(t, err)
	defer conn.Close()

	reader := bufio.NewReader(conn)

	// Oversized message → error response.
	_, err = fmt.Fprint(conn, strings.Repeat("x", 20)+"\n")
	require.NoError(t, err)

	line, err := reader.ReadString('\n')
	require.NoError(t, err)
	require.Equal(t, "[error] message too large\n", line)

	// Normal message on the same connection should still work.
	_, err = fmt.Fprintln(conn, "ok")
	require.NoError(t, err)

	line, err = reader.ReadString('\n')
	require.NoError(t, err)
	require.Equal(t, "ok\n", line)
}

func TestHandleQueries_ConcurrentConnections(t *testing.T) {
	t.Parallel()

	const numConnections = 20

	addr := startServer(t, defaultCfg(), echoHandler)

	var wg sync.WaitGroup
	wg.Add(numConnections)

	for i := 0; i < numConnections; i++ {
		i := i
		go func() {
			defer wg.Done()

			conn, err := net.Dial("tcp", addr)
			if err != nil {
				t.Errorf("connection %d: dial error: %v", i, err)
				return
			}
			defer conn.Close()

			msg := fmt.Sprintf("msg-%d", i)
			_, err = fmt.Fprintln(conn, msg)
			if err != nil {
				t.Errorf("connection %d: write error: %v", i, err)
				return
			}

			line, err := bufio.NewReader(conn).ReadString('\n')
			if err != nil {
				t.Errorf("connection %d: read error: %v", i, err)
				return
			}

			if line != msg+"\n" {
				t.Errorf("connection %d: got %q, want %q", i, line, msg+"\n")
			}
		}()
	}

	wg.Wait()
}

func TestHandleQueries_Ping(t *testing.T) {
	t.Parallel()

	addr := startServer(t, defaultCfg(), echoHandler)

	conn, err := net.Dial("tcp", addr)
	require.NoError(t, err)
	defer conn.Close()

	reader := bufio.NewReader(conn)

	// PING is case-insensitive and handled at the network layer, not by the app handler.
	for _, msg := range []string{"PING", "ping", "Ping"} {
		_, err = fmt.Fprintln(conn, msg)
		require.NoError(t, err)

		line, err := reader.ReadString('\n')
		require.NoError(t, err)
		require.Equal(t, "PONG\n", line, "input: %q", msg)
	}
}

func TestHandleQueries_IdleTimeout(t *testing.T) {
	t.Parallel()

	cfg := defaultCfg()
	cfg.IdleTimeoutDuration = 150 * time.Millisecond

	addr := startServer(t, cfg, echoHandler)

	conn, err := net.Dial("tcp", addr)
	require.NoError(t, err)
	defer conn.Close()

	// Do not send anything; server must close the connection after idle timeout.
	conn.SetDeadline(time.Now().Add(2 * time.Second))
	buf := make([]byte, 1)
	_, err = conn.Read(buf)
	require.Error(t, err, "connection should be closed by server after idle timeout")
}

func TestHandleQueries_PanicRecovery(t *testing.T) {
	t.Parallel()

	panicHandler := func(_ context.Context, _ []byte) []byte {
		panic("test panic")
	}

	addr := startServer(t, defaultCfg(), panicHandler)

	// Handler panics — server must recover, close the connection, and keep running.
	conn1, err := net.Dial("tcp", addr)
	require.NoError(t, err)
	defer conn1.Close()

	_, err = fmt.Fprintln(conn1, "SET x 1")
	require.NoError(t, err)

	buf := make([]byte, 1)
	_, err = conn1.Read(buf)
	require.Error(t, err, "connection should be closed after panic recovery")

	// Server must still accept new connections after the panic.
	conn2, err := net.Dial("tcp", addr)
	require.NoError(t, err)
	conn2.Close()
}
