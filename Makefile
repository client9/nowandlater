
test: lint
	go test ./...

clean:
	rm -f nldate
	rm -f cover.out coverage.html coverage.out cover.out.tmp
	rm -rf testdata
	rm -f nowandlater.test

fmt:
	gofmt -w -s .

lint: fmt
	go vet ./...
	go fix -diff .
	golangci-lint run ./...

# test but ignore any fuzz stuff
cover:
	rm -f cover.out
	go test -run='^Test' -coverprofile=cover.out -coverpkg=.,./languages,./internal/engine ./...
	grep -v '/cmd/' cover.out > cover.out.tmp && mv cover.out.tmp cover.out
	go tool cover -func=cover.out

fuzz:
	cd tests && go test -fuzz=.

bench:
	go test -bench=. -benchmem -benchtime=3s ./...
