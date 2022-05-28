add: check build test

.PHONY: acctest
acctest:
	@TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

test:
	@go test ./... -v -race -vet=off $(TESTARGS)

check:
	@go vet ./...
	@go generate ./...

build:
	@go build ./...

