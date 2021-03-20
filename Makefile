all: test build

test:
	go test -v ./...

format:
	npm run format
	go fmt ./...

deps: js-deps go-deps

js-deps:
	npm install

go-deps:
	go mod tidy

js: js-deps
	npm run build

build: go-deps js
	go build -o bin/jabberwocky

release: go-deps js
	go build -ldflags "-s -w" -o bin/jabberwocky
