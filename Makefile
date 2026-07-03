.PHONY: fmt vet test run init clean

fmt:
	gofmt -w .

vet:
	go vet ./...

test:
	go test ./...

run:
	go run ./cmd/toychain help

init:
	go run ./cmd/toychain init -difficulty 2

clean:
	rm -f data/chain.json