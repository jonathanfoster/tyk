SHELL := /bin/bash

GOCMD=go
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOBUILD=$(GOCMD) build
GOINSTALL=$(GOCMD) install

BINARY_NAME=tyk
BINARY_LINUX=tyk
BUILD_PLATFORM=linux/amd64
TAGS=coprocess grpc goplugin
CONF=tyk.conf

TEST_REGEX=.
TEST_COUNT=1

BENCH_REGEX=.
BENCH_RUN=NONE

DOCKER_CMD?=docker

.PHONY: test
test:
	$(GOTEST) -run=$(TEST_REGEX) -count=$(TEST_COUNT) ./...

# lint runs all local linters that must pass before pushing
.PHONY: lint lint-install lint-fast
lint:
	task lint

.PHONY: bench
bench:
	$(GOTEST) -run=$(BENCH_RUN) -bench=$(BENCH_REGEX) ./...

.PHONY: clean
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

.PHONY: dev
dev:
	$(GOBUILD) -tags "$(TAGS)" -o $(BINARY_NAME) -v .
	./$(BINARY_NAME) --conf $(CONF)

.PHONY: build
build:
	$(GOBUILD) -tags "$(TAGS)" -o $(BINARY_NAME) -trimpath .

.PHONY: build-linux
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -tags "$(TAGS)" -o $(BINARY_LINUX) -v .

.PHONY: install
install:
	$(GOINSTALL) -tags "$(TAGS)"

.PHONY: install-deps
install-deps:
	go install github.com/go-task/task/v3/cmd/task@latest
	task deps

.PHONY: db-start
db-start: redis-start mongo-start

.PHONY: db-stop
db-stop: redis-stop mongo-stop

# Docker start redis
.PHONY: redis-start
redis-start:
	$(DOCKER_CMD) run -itd --rm --name redis -p 127.0.0.1:6379:6379 redis:4.0-alpine redis-server --appendonly yes

.PHONY: redis-stop
redis-stop:
	$(DOCKER_CMD) stop redis

.PHONY: redis-cli
redis-cli:
	$(DOCKER_CMD) exec -it redis redis-cli

# Docker start mongo
.PHONY: mongo-start
mongo-start:
	$(DOCKER_CMD) run -itd --rm --name mongo -p 127.0.0.1:27017:27017 mongo:3.4-jessie

.PHONY: mongo-stop
mongo-stop:
	$(DOCKER_CMD) stop mongo

.PHONY: mongo-shell
mongo-shell:
	$(DOCKER_CMD) exec -it mongo mongo

.PHONY: docker
docker:
	$(DOCKER_CMD) build --platform ${BUILD_PLATFORM} --rm -t internal/tyk-gateway .

.PHONY: docker-std
docker-std: build
	$(DOCKER_CMD) build --platform ${BUILD_PLATFORM} --no-cache -t internal/tyk-gateway:std -f ci/Dockerfile.std .

