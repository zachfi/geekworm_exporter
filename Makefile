.PHONY: all build compile-only test test-unit test-integration lint tools release-publish clean

BINARY   := geekworm_exporter
CMD      := ./cmd/$(BINARY)
GOFLAGS  := -mod=vendor

all: build

build:
	go build $(GOFLAGS) -o $(BINARY) $(CMD)

# Cross-compile for Raspberry Pi (armv7); matches goreleaser targets.
compile-only:
	GOARCH=arm GOARM=7 GOOS=linux CGO_ENABLED=0 go build $(GOFLAGS) -o /dev/null $(CMD)

test: test-unit

test-unit:
	go test $(GOFLAGS) -v ./...

test-integration:
	@echo "No integration tests defined (hardware required)"

lint:
	go vet $(GOFLAGS) ./...

tools:
	@echo "No extra tools required"

release-publish:
	goreleaser release --clean

clean:
	rm -f $(BINARY)
