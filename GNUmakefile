add: check build test

.PHONY: acctest
acctest:
	@tools/pretest.sh

test:
	@go test ./... -v -race -vet=off $(TESTARGS)

check:
	@go vet ./...
	@go generate ./...

build:
	@go build ./...

