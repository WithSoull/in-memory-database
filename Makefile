-include .env

LOCAL_BIN=$(CURDIR)/bin

.PHONY: get-deps install-deps test test-cover

get-deps:
	go get -u go.uber.org/zap
	go get -u github.com/stretchr/testify/require
	go get -u github.com/gojuno/minimock/v3

install-deps:
		GOBIN=$(LOCAL_BIN) go install github.com/gojuno/minimock/v3/cmd/minimock@v3.3.5

test:
	go test ./...

test-cover:
	go test ./... -coverpkg=./... -coverprofile=coverage.out
	go tool cover -func=coverage.out | tail -n 1
