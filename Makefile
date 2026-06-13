.PHONY: lint lint-install test test-short test-race vet docker-build release
LINT_VERSION ?= v2.12.2
GOLANGCI_LINT ?= $(shell command -v golangci-lint 2>/dev/null || printf '%s/bin/golangci-lint' "$$(go env GOPATH)")

lint:
	@echo "Running gofmt check..."
	@test -z "$(gofmt -l .)" || (gofmt -d . && exit 1)
	@test -x "$(GOLANGCI_LINT)" || { \
		echo "golangci-lint not found. Run: make lint-install"; \
		exit 1; \
	}
	@$(GOLANGCI_LINT) run ./...

lint-install:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(LINT_VERSION)

vet:
	go vet ./...

test:
	go test ./...

test-short:
	go test ./... -short

test-race:
	go test -race ./...

docker-build:
	docker build -t emuready-discord-giveaway:local .

docker-run:
	docker-compose up --build

release:
	echo "CI handles release; run from pipeline."
