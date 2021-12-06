all: build

build:
	echo "Building..."
	go build -o bin/bot ./cmd/...

vendor:
	go mod vendor