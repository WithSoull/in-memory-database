package network

import "context"

type TCPHandler = func(context.Context, []byte) []byte

type TCPServer interface {
	HandleQueries(ctx context.Context, handler TCPHandler)
}

