BINARY  = granola
INSTALL = $(shell go env GOPATH)/bin

.PHONY: build clean install test

build:
	go build -o $(BINARY) .

clean:
	rm -f $(BINARY)

install:
	go install .

test:
	go test ./...
