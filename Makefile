SHELL := sh

.PHONY: help
.DEFAULT_GOAL := help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: build
build: ## build
	go build ./...

.PHONY: test
test: ## test
	go test ./...

.PHONY: clean
clean: ## cleanup
	rm -f nldate
	rm -f cover.out coverage.html coverage.out cover.out.tmp
	rm -rf testdata
	rm -f nowandlater.test

.PHONY: lintverify
## NOTE: this downloads it's schema over the network
lintverify:
	golangci-lint config verify

.PHONY: fmt
fmt: ## reformat source code
	go mod tidy
	gofmt -w -s *.go

.PHONY: lint
lint: ## lint and verify repo is already formatted
	go mod tidy
	git diff --exit-code -- go.mod go.sum
	golangci-lint run .

.PHONY: cover
cover: ## coverage, no fuzz
	rm -f cover.out
	go test -run='^Test' -coverprofile=cover.out -coverpkg=.,./languages,./internal/engine ./...
	grep -v '/cmd/' cover.out > cover.out.tmp && mv cover.out.tmp cover.out
	go tool cover -func=cover.out

.PHONY: fuzz
fuzz: ## fuzz test
	cd tests && go test -fuzz=.

.PHONY: bench
bench: ## benchmarks
	go test -bench=. -benchmem -benchtime=3s ./...
