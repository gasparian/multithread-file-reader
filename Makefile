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
		-v -o ./cmd/filereader \
		./cmd

.PHONY: run
run:
	./cmd/filereader

.PHONY: test
test:
	time go test -v -cover -race -count=1 -timeout 30s $$(go list ./... | grep -v /cmd)