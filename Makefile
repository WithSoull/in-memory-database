include .env

LOCAL_BIN=$(CURDIR)/bin

get-deps:
	go get -u go.uber.org/zap
	go get -u github.com/stretchr/testify/require
	go get -u github.com/gojuno/minimock/v3

install-deps:
		GOBIN=$(LOCAL_BIN) go install github.com/gojuno/minimock/v3/cmd/minimock@v3.3.5
