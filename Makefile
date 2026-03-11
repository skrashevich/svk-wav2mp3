BINARY     := wav2mp3
CMD        := ./cmd/wav2mp3
BIN_DIR    := ./bin
VERSION    := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS    := -ldflags "-X main.version=$(VERSION)"

.PHONY: all build clean install test fmt lint

all: build

build:
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 go build $(LDFLAGS) -o $(BIN_DIR)/$(BINARY) $(CMD)
	@echo "Built: $(BIN_DIR)/$(BINARY) (version: $(VERSION))"

install: build
	cp $(BIN_DIR)/$(BINARY) /usr/local/bin/$(BINARY)
	@echo "Installed to /usr/local/bin/$(BINARY)"

clean:
	rm -rf $(BIN_DIR)

test:
	CGO_ENABLED=0 go test ./... -v -count=1

fmt:
	go fmt ./...

lint:
	golangci-lint run ./...

testdata:
	go run testdata/gen_fixtures.go
