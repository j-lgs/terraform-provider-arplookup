add: build test

.PHONY: acctest
acctest:
	@tools/pretest.sh

test:
	@echo "generate ---> testing code"
	@go test ./... -v -race -vet=off -$(TESTARGS)

generate:
	@echo "generate ---> generating code"
	@go generate ./...

check:
	@echo "check ---> linting source code"
	@go vet ./...
	@#go run github.com/golangci/golangci-lint/cmd/golangci-lint run ./internal/arplookup
	@go run honnef.co/go/tools/cmd/staticcheck ./internal/arplookup

build: check
	@echo "build ---> building code"
	@go build -v .

