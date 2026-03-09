package tcpserver

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	appconfig "github.com/WithSoull/in-memory-database/internal/config/app"
	derrors "github.com/WithSoull/in-memory-database/internal/domainerrors"
	"github.com/WithSoull/in-memory-database/internal/network"
	"go.uber.org/zap"
)

type server struct {
	listener       net.Listener
	idleTimeout    time.Duration
	maxMessageSize int
	semaphore      chan struct{}
	logger         *zap.Logger
}

func NewServer(cfg appconfig.NetworkConfig, logger *zap.Logger) (network.TCPServer, error) {
	if logger == nil {
		return nil, derrors.ErrInvalidLogger
	}

	listener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %w", err)
	}

	return &server{
		listener:       listener,
		idleTimeout:    cfg.IdleTimeoutDuration,
		maxMessageSize: cfg.MaxMessageSizeBytes,
		semaphore:      make(chan struct{}, cfg.MaxConnections),
		logger:         logger,
	}, nil
}

func (s *server) HandleQueries(ctx context.Context, handler network.TCPHandler) {
	go func() {
		<-ctx.Done()
		s.listener.Close()
	}()

	wg := sync.WaitGroup{}

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				break
			}
			s.logger.Error("failed to accept connection", zap.Error(err))
			continue
		}

		select {
		case s.semaphore <- struct{}{}:
		default:
			s.logger.Warn("max connections reached, rejecting connection",
				zap.String("remote_addr", conn.RemoteAddr().String()),
			)
			conn.Close()
			continue
		}

		wg.Add(1)
		go func(connection net.Conn) {
			defer wg.Done()
			defer func() { <-s.semaphore }()
			s.handleConnection(ctx, connection, handler)
		}(conn)
	}

	wg.Wait()
}

func (s *server) handleConnection(ctx context.Context, conn net.Conn, handler network.TCPHandler) {
	defer func() {
		if r := recover(); r != nil {
			s.logger.Error("panic in connection handler", zap.Any("recover", r))
		}
		conn.Close()
	}()

	// Unblock any pending read when the server context is cancelled.
	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-ctx.Done():
			conn.SetDeadline(time.Now())
		case <-done:
		}
	}()

	reader := newBufioReader(conn, s.maxMessageSize+1)

	for {
		if ctx.Err() != nil {
			return
		}

		if s.idleTimeout > 0 {
			conn.SetDeadline(time.Now().Add(s.idleTimeout))
		}

		line, isPrefix, err := reader.ReadLine()
		if err != nil {
			if !isTimeout(err) && !errors.Is(err, net.ErrClosed) {
				s.logger.Debug("connection read error", zap.Error(err))
			}
			return
		}

		if ctx.Err() != nil {
			return
		}

		if isPrefix {
			if err := drainLine(reader); err != nil {
				return
			}
			_, _ = conn.Write([]byte("[error] message too large\n"))
			continue
		}

		if bytes.EqualFold(line, []byte("PING")) {
			_, _ = conn.Write([]byte("PONG\n"))
			continue
		}

		response := handler(ctx, line)
		response = append(response, '\n')
		_, _ = conn.Write(response)
	}
}

func newBufioReader(conn net.Conn, size int) *bufio.Reader {
	return bufio.NewReaderSize(conn, size)
}

func isTimeout(err error) bool {
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}

func drainLine(reader interface{ ReadLine() ([]byte, bool, error) }) error {
	for {
		_, isPrefix, err := reader.ReadLine()
		if err != nil {
			return err
		}
		if !isPrefix {
			return nil
		}
	}
}
