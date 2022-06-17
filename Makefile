export DOCKER_BUILDKIT=1
GIT_COMMIT_FULL  := $(shell git rev-parse HEAD)
DOCKER_REGISTRY  := "quay.io/packet/ironlib"
LINTER_EXPECTED_VERSION   := "1.46.2"

.DEFAULT_GOAL := help

## lint
lint:
	(golangci-lint --version | grep -q "${LINTER_EXPECTED_VERSION}" && golangci-lint run --config .golangci.yml) \
		|| echo "expected linter version: ${LINTER_EXPECTED_VERSION}"

## Go test
test:
	CGO_ENABLED=0 go test -v -covermode=atomic ./...

## build docker image and tag as quay.io/packet/ironlib:latest
build-image:
	@echo ">>>> NOTE: You may want to execute 'make build-image-nocache' depending on the Docker stages changed"
	docker build --rm=true -f Dockerfile -t ${DOCKER_REGISTRY}:latest  . \
							 --label org.label-schema.schema-version=1.0 \
							 --label org.label-schema.vcs-ref=$(GIT_COMMIT_FULL) \
							 --label org.label-schema.vcs-url=https://github.com/metal-toolbox/ironlib.git

## build docker image, ignoring the cache
build-image-nocache:
	docker build --no-cache --rm=true -f Dockerfile -t ${DOCKER_REGISTRY}:latest  . \
							 --label org.label-schema.schema-version=1.0 \
							 --label org.label-schema.vcs-ref=$(GIT_COMMIT_FULL) \
							 --label org.label-schema.vcs-url=https://github.com/metal-toolbox/ironlib.git


## push docker image
push-image:
	docker push ${DOCKER_REGISTRY}:latest



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
