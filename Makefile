BINARY_NAME=mytmbar
INSTALL_PATH=/usr/local/bin/$(BINARY_NAME)

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
LDFLAGS := -X main.version=$(VERSION) -X main.commit=$(COMMIT)

.PHONY: build
build:
	go build -ldflags "$(LDFLAGS)" .

.PHONY: install
install: build
	@sudo cp ./$(BINARY_NAME) $(INSTALL_PATH)
	@sudo chmod +x $(INSTALL_PATH)

.PHONY: test
test: build
	go test -v ./...

.PHONY: lint
lint:
	golangci-lint run

.PHONY: clean
clean:
	go clean
