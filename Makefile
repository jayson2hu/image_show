.PHONY: dev build build-frontend build-all test docker-build docker-up docker-down

dev:
	air

build:
	CGO_ENABLED=0 go build -o image-show .

build-frontend:
	cd web && pnpm install && pnpm build

build-all: build-frontend build

test:
	go test ./...

docker-build:
	docker build -t image-show .

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down
