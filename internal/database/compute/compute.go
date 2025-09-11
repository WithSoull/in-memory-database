package compute

type Query interface {
	CommandID() int64
	Arguments() []string
}

type Compute interface {
	Parse(string) (Query, error)
}
