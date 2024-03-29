BINPATH ?= build

MY_UID=$(shell id -u)
MY_GID=$(shell id -g)
BUILD_TIME=$(shell date +%s)
GIT_COMMIT=$(shell git rev-parse HEAD)
VERSION ?= $(shell git tag --points-at HEAD | grep ^v | head -n 1)

LDFLAGS = -ldflags "-X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT) -X main.Version=$(VERSION)"

.PHONY: all
all: audit test build

.PHONY: audit
audit:
	go list -m all | nancy sleuth

.PHONY: build
build:
	go build -tags 'production' $(LDFLAGS) -o $(BINPATH)/dp-import-cantabular-dataset

.PHONY: debug
debug:
	go build -tags 'debug' $(LDFLAGS) -o $(BINPATH)/dp-import-cantabular-dataset
	HUMAN_LOG=1 DEBUG=1 $(BINPATH)/dp-import-cantabular-dataset

.PHONY: debug-run
debug-run:
	HUMAN_LOG=1 DEBUG=1 go run -race -tags 'debug' $(LDFLAGS) main.go

.PHONY: test
test:
	go test -race -cover ./...

.PHONY: produce
produce:
	HUMAN_LOG=1 go run cmd/producer/main.go

.PHONY: convey
convey:
	goconvey ./...

.PHONY: test-component
test-component:
	cd features/compose && MY_UID=$(MY_UID) MY_GID=$(MY_GID) docker-compose up --build --abort-on-container-exit
	cd features/compose && docker-compose down --volume
	echo "please ignore error codes 0, like so: ERROR[xxxx] 0, as error code 0 means that there was no error"

.PHONY: fmt
fmt:
	go fmt ./...
.PHONY: lint
lint:
	exit
