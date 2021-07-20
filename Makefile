BINPATH ?= build

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
	HUMAN_LOG=1 DEBUG=1 go run -tags 'debug' $(LDFLAGS) main.go

.PHONY: test
test:
	go test -v -race -cover ./...

.PHONY: produce
produce:
	HUMAN_LOG=1 go run cmd/producer/main.go

.PHONY: convey
convey:
	goconvey ./...

.PHONY: test-component
test-component:
	cd features/steps; docker-compose up --build --abort-on-container-exit; \
	CODE=$$; \
	docker-compose down --volume
	cat logs && rm logs
	exit $$CODE

.PHONY: fmt
fmt:
	go fmt ./...
.PHONY: lint
lint:
	exit
