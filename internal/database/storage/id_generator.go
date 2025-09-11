package storage

type IDGenerator struct {
	counter int64
}

func NewIDGenerator() *IDGenerator {
	generator := &IDGenerator{counter: 0}
	return generator
}

func (g *IDGenerator) Generate() int64 {
	g.counter += 1
	return g.counter
}
