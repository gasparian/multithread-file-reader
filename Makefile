.SILENT:

.DEFAULT_GOAL := build

ifndef GOOS
	export GOOS=darwin
endif
ifndef GOARCH
	export GOARCH=amd64
endif

.PHONY: install-hooks
install-hooks:
	cp -rf .githooks/pre-commit .git/hooks/pre-commit

.PHONY: build
build:
	go build \
	    -ldflags '-w -extldflags "-static"' \
		-v -o ./cmd/filereader/filereader \
		./cmd/filereader

.PHONY: run
run:
	./cmd/filereader/filereader

.PHONY: test
test:
	CGO_ENABLED=1 go test -v -cover -race -count=1 -timeout 30s $$(go list ./... | grep -v '/cmd')

.PHONY: perftest
perftest:
	go run ./cmd/perf/main.go