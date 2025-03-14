MOD_PATHS  := $(wildcard ./*mod/)
MOD_NAMES  := $(MOD_PATHS:./%/=%)
MODS_TIDY  := $(MOD_NAMES:%=tidy/%)
MODS_TEST  := $(MOD_NAMES:%=test/%)
MODS_LINT  := $(MOD_NAMES:%=lint/%)
MODS_TOOLS := $(MOD_NAMES:%=tools/%)

.PHONY: all
all: clean .WAIT lint test

.PHONY: lint
lint: ${MODS_LINT} ## Run linter

.PHONY: ${MODS_LINT}
${MODS_LINT}:
	go tool github.com/golangci/golangci-lint/cmd/golangci-lint run --timeout=15m ./${@F}/...

.PHONY: test
test: ${MODS_TEST} ## Run tests

.PHONY: ${MODS_TEST}
${MODS_TEST}:
	mkdir -p .output/coverage .output/junit
	go tool gotest.tools/gotestsum --junitfile=.output/junit/${@F}.xml -- -race -covermode=atomic -coverprofile=.output/coverage/${@F}.txt ./${@F}/...

.PHONY:tidy ## Tidy all mods
tidy: ${MODS_TIDY}

.PHONY: ${MODS_TIDY}
${MODS_TIDY}:
	cd ./${@F} && go mod tidy

.PHONY: tools
tools: ${MODS_TOOLS} ## Update tools

.PHONY: ${MODS_TOOLS}
${MODS_TOOLS}:
	cd ./${@F} && go get -tool gotest.tools/gotestsum@latest
	cd ./${@F} && go get -tool github.com/golangci/golangci-lint/cmd/golangci-lint@latest

.PHONY: clean
clean: ## Clean files
	git clean -Xdf
