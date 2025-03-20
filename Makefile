.DEFAULT_GOAL := help

export GOBIN=$(CURDIR)/bin
export PATH:=$(PATH):$(GOBIN)

## Run all linters
lint: golangci-lint check-go-generated

## Run golangci-lint
golangci-lint:
	golangci-lint run --config .golangci.yml

## Run go generate
generate:
	go install golang.org/x/tools/cmd/stringer@v0.31.0
	go generate ./...

## Check generated files are up to date
check-go-generated: generate
	git diff | (! grep .)

## Run go test
go-test:
	CGO_ENABLED=0 go test -v -covermode=atomic ./...

## Run all tests and linters
test: go-test lint

# https://gist.github.com/prwhite/8168133
# COLORS
GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
WHITE  := $(shell tput -Txterm setaf 7)
RESET  := $(shell tput -Txterm sgr0)


TARGET_MAX_CHAR_NUM=20
## Show help
help:
	@echo ''
	@echo 'Usage:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target>${RESET}'
	@echo ''
	@echo 'Targets:'
	@awk '/^[a-zA-Z\-\\_0-9]+:/ { \
		helpMessage = match(lastLine, /^## (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")-1); \
			helpMessage = substr(lastLine, RSTART + 3, RLENGTH); \
			printf "  ${YELLOW}%-$(TARGET_MAX_CHAR_NUM)s${RESET} ${GREEN}%s${RESET}\n", helpCommand, helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)
