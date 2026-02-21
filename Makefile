BINARY     := wav2mp3
CMD        := ./cmd/wav2mp3
BIN_DIR    := ./bin
VERSION    := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS    := -ldflags "-X main.version=$(VERSION)"
# Homebrew arm64 (Apple Silicon) путь
HOMEBREW_PREFIX := $(shell brew --prefix 2>/dev/null || echo /opt/homebrew)
CGO_CFLAGS := -I$(HOMEBREW_PREFIX)/include
CGO_LDFLAGS:= -L$(HOMEBREW_PREFIX)/lib -lmp3lame

.PHONY: all build clean install test fmt lint

all: build

build:
	@mkdir -p $(BIN_DIR)
	CGO_CFLAGS="$(CGO_CFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS)" \
		go build $(LDFLAGS) -o $(BIN_DIR)/$(BINARY) $(CMD)
	@echo "Собран: $(BIN_DIR)/$(BINARY) (версия: $(VERSION))"

install: build
	cp $(BIN_DIR)/$(BINARY) /usr/local/bin/$(BINARY)
	@echo "Установлен в /usr/local/bin/$(BINARY)"

clean:
	rm -rf $(BIN_DIR)

test:
	CGO_CFLAGS="$(CGO_CFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS)" \
		go test ./... -v -count=1

fmt:
	go fmt ./...

lint:
	golangci-lint run ./...

testdata:
	go run testdata/gen_fixtures.go
