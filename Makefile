.SILENT:

.DEFAULT_GOAL := build

.PHONY: install-hooks
install-hooks:
	cp -rf .githooks/pre-commit .git/hooks/pre-commit

.PHONY: build
build:
	go build \
	    -ldflags '-w -extldflags "-static"' \
		-v -o ./filereader \
		./cmd/filereader

.PHONY: test
test:
	go test -v -cover -race -count=1 -timeout 30s $$(go list ./... | grep -v '/cmd')

.PHONY: perftest
perftest:
	go run ./cmd/perf/main.go