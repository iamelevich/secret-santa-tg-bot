all: build

docker_push:
	docker push beer13/secret-santa-tg-bot:latest

docker:
	echo "Building docker image..."
	docker compose -f docker/docker-compose.yml build

build:
	echo "Building..."
	go build -o bin/bot ./cmd/...

vendor:
	echo "Updating vendors..."
	go mod vendor