SHELL := sh

.PHONY: help
.DEFAULT_GOAL := help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## build
	go build ./...

test: ## test
	go test ./...

clean: ## cleanup
	rm -f nldate
	rm -f cover.out coverage.html coverage.out cover.out.tmp
	rm -rf testdata
	rm -f nowandlater.test

## NOTE: this downloads it's schema over the network
lintverify:
	golangci-lint config verify

fmt: ## reformat source code
	go mod tidy
	gofmt -w -s *.go

lint: ## lint and verify repo is already formatted
	go mod tidy
	git diff --exit-code -- go.mod go.sum
	test -z "$$(gofmt -l *.go)"
	golangci-lint run .

cover: ## coverage, no fuzz
	rm -f cover.out
	go test -run='^Test' -coverprofile=cover.out -coverpkg=.,./languages,./internal/engine ./...
	grep -v '/cmd/' cover.out > cover.out.tmp && mv cover.out.tmp cover.out
	go tool cover -func=cover.out

fuzz: ## fuzz test
	cd tests && go test -fuzz=.

bench: ## benchmarks
	go test -bench=. -benchmem -benchtime=3s ./...
