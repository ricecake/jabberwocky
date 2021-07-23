all: format lint compile test

compile: deps build

quick: js-build go-build

test:
	go test -v ./...

format: go-fmt js-fmt

js-fmt:
	npm run format

go-fmt:
	go fmt ./...

lint: go-lint js-lint

go-lint:
	go vet

js-lint:
	echo "nope"

deps: js-deps go-deps

js-deps:
	npm install

go-deps:
	go mod tidy

js: js-deps js-build

js-build:
	npm run build

go: go-deps go-build

go-build:
	go build -o bin/jabberwocky

build: js-build go-build

release: go-deps js
	go build -ldflags "-s -w" -o bin/jabberwocky
