.PHONY: build test
build:
	go build -trimpath -o bin/pass-gen .
test:
	go test ./...
