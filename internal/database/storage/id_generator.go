package storage

import "sync/atomic"

type IDGenerator struct {
	counter int64
}

func NewIDGenerator() *IDGenerator {
	generator := &IDGenerator{counter: 0}
	return generator
}

func (g *IDGenerator) Generate() int64 {
	return atomic.AddInt64(&g.counter, 1)
}
