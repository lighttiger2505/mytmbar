BINARY_NAME=mytmbar
INSTALL_PATH=/usr/local/bin/$(BINARY_NAME)

.PHONY: build
build:
	go build .

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
