BINARY_NAME=text-to-sql-proxy
DIST_DIR=dist
CMD_PATH=./src/cmd/text-to-sql-proxy

VERSION ?= dev
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS = -ldflags "-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildDate=$(BUILD_DATE)"

.PHONY: all build build-all build-windows build-linux build-darwin-arm64 test clean

all: test build

build:
	@mkdir -p $(DIST_DIR)
	go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME) $(CMD_PATH)

build-all: build-windows build-linux build-darwin-arm64

build-windows:
	@mkdir -p $(DIST_DIR)
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe $(CMD_PATH)

build-linux:
	@mkdir -p $(DIST_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 $(CMD_PATH)

build-darwin-arm64:
	@mkdir -p $(DIST_DIR)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 $(CMD_PATH)

test:
	go test -v ./...

clean:
	rm -rf $(DIST_DIR)
	go clean
