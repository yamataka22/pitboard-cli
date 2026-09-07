.PHONY: build test install docs release

DEV_VERSION := dev-$(shell git rev-parse --short HEAD 2>/dev/null || echo local)
LDFLAGS := -X main.version=$(DEV_VERSION)

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

# リリース。README のインストール手順にあるファイル名の版を更新してコミットし、タグを打って push する。
# タグ push で GitHub Actions の goreleaser が Releases を作る（README の版が合っていなければそこで止まる）
release:
	@test -n "$(VERSION)" || { echo "usage: make release VERSION=0.1.1"; exit 1; }
	@git diff --quiet && git diff --cached --quiet || { echo "commit or stash your changes first"; exit 1; }
	$(MAKE) test
	goreleaser check
	sed -i.bak -E 's/pitboard_[0-9]+\.[0-9]+\.[0-9]+_/pitboard_$(VERSION)_/g; s#/download/v[0-9]+\.[0-9]+\.[0-9]+/#/download/v$(VERSION)/#g' README.md && rm README.md.bak
	git diff --quiet README.md || git commit -am "v$(VERSION)"
	git tag -a v$(VERSION) -m "v$(VERSION)"
	git push origin main v$(VERSION)
