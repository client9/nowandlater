
clean:
	rm -f nldate
	rm -f cover.out
	rm -rf testdata

fmt:
	gofmt -w -s .

lint: fmt
	go vet ./...
	go fix -diff .
	golangci-lint run ./...

test:
	go vet
	go test ./...

# test but ignore any fuzz stuff
cover:
	rm -f cover.out
	go test -run='^Test' -coverprofile=cover.out
	go tool cover -func=cover.out

fuzz:
	go test -fuzz=.

bench:
	go test -bench=. -benchmem -benchtime=3s ./...
