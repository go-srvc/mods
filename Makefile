.PHONY: test clean

MOD_PATHS := $(wildcard ./*mod/)
MOD_NAMES := $(MOD_PATHS:./%/=%)
MODS_TIDY := $(MOD_NAMES:%=tidy/%)
MODS_TEST := $(MOD_NAMES:%=test/%)
MODS_LINT := $(MOD_NAMES:%=lint/%)
MODS_TOOLS := $(MOD_NAMES:%=tools/%)

all: clean lint test

.PHONY: lint
lint: ${MODS_LINT} ## Run linter

.PHONY: ${MODS_LINT}
${MODS_LINT}:
	cd ./${@F} && go tool github.com/golangci/golangci-lint/cmd/golangci-lint run --timeout=15m ./...

.PHONY: test
test: ${MODS_TEST} ## Run tests

.PHONY: ${MODS_TEST}
${MODS_TEST}:
	cd ./${@F} && go tool gotest.tools/gotestsum --junitfile=junit.xml -- -race -covermode=atomic -coverprofile=coverage.txt ./...

.PHONY:tidy ## tidy all mods
tidy: ${MODS_TIDY}

.PHONY: ${MODS_TIDY}
${MODS_TIDY}:
	cd ./${@F} && go mod tidy

.PHONY: tools
tools: ${MODS_TOOLS} ## Install tools

.PHONY: ${MODS_TOOLS}
${MODS_TOOLS}:
	cd ./${@F} && go get -tool gotest.tools/gotestsum@latest
	cd ./${@F} && go get -tool github.com/golangci/golangci-lint/cmd/golangci-lint@latest

.PHONY: clean
clean: ## Clean files
	git clean -Xdf
