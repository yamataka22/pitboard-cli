.PHONY: build test install docs

VERSION := dev-$(shell git rev-parse --short HEAD 2>/dev/null || echo local)
LDFLAGS := -X main.version=$(VERSION)

build:
	go build -ldflags "$(LDFLAGS)" -o pitboard ./cmd/pitboard

test:
	test -z "$$(gofmt -l .)" && go vet ./... && go test ./...

# go の bin ディレクトリに pitboard を入れる。mise 管理の Go なら shim も作り直す
install:
	go install -ldflags "$(LDFLAGS)" ./cmd/pitboard
	@command -v mise >/dev/null 2>&1 && mise reshim || true
	@echo "installed to $$(go env GOBIN 2>/dev/null || echo $$(go env GOPATH)/bin)/pitboard"

# docs/commands/ を cobra の定義から生成し直す
docs:
	go run ./cmd/gendocs docs/commands
