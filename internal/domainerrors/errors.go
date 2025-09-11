package derrors

import "errors"

var (
	ErrIvalidEngine  = errors.New("engine is invalid")
	ErrInvalidLogger = errors.New("logger is invalid")
)

var (
	ErrInvalidQuery     = errors.New("empty query")
	ErrInvalidCommand   = errors.New("invalid command")
	ErrInvalidArguments = errors.New("invalid arguments")
)

var (
	ErrKeyNotFound = errors.New("key not found")
)
