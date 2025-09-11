package parser

import "github.com/WithSoull/in-memory-database/internal/database/compute"

type Query struct {
	commandID int64
	arguments []string
}

func NewQuery(commandID int64, arguments []string) compute.Query {
	return &Query{
		commandID: commandID,
		arguments: arguments,
	}
}

func (c *Query) CommandID() int64 {
	return c.commandID
}

func (c *Query) Arguments() []string {
	return c.arguments
}
