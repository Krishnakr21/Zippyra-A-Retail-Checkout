.PHONY: all test build-windows build-linux build-darwin clean

BINARY_NAME=zippyra-connector

all: test build-windows build-linux build-darwin

test:
	go test -v ./...

build-windows:
	mkdir -p bin
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o bin/$(BINARY_NAME).exe .

build-linux:
	mkdir -p bin
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/$(BINARY_NAME)-linux .

build-darwin:
	mkdir -p bin
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o bin/$(BINARY_NAME)-darwin .

clean:
	rm -rf bin/
