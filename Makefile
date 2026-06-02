MOD_PATHS     := $(wildcard ./*mod/)
MOD_NAMES     := ${MOD_PATHS:./%/=%}
MODS_TIDY     := ${MOD_NAMES:%=tidy/%}
MODS_CHECK    := ${MOD_NAMES:%=tidy-check/%}
MODS_TEST     := ${MOD_NAMES:%=test/%}
MODS_LINT     := ${MOD_NAMES:%=lint/%}
MODS_UPDATE   := ${MOD_NAMES:%=update-deps/%}
MODS_DOWNLOAD := ${MOD_NAMES:%=download/%}
MODS_APICHECK := ${MOD_NAMES:%=api-check/%}
MODS_TAG      := ${MOD_NAMES:%=tag/%}
MODS_CROSS    := ${MOD_NAMES:%=cross-check/%}

TOOLS_DIR     := internal/tools
BIN           := $(abspath .bin)
GOLANGCILINT  := ${BIN}/golangci-lint
GOTESTSUM     := ${BIN}/gotestsum
GORELEASE     := ${BIN}/gorelease

REMOTE        ?= origin
mod_last_tag   = $(shell git ls-remote --tags --refs ${REMOTE} '$(1)/v*' | sed 's|.*refs/tags/||' | sort -V | tail -1)

.PHONY: all
all: clean tidy-check .WAIT lint test

.PHONY: tools
tools: ## Build dev tools
	cd ${TOOLS_DIR} && GOBIN=${BIN} go install tool

.NOTPARALLEL: lint
.PHONY: lint
lint: ${MODS_LINT} ## Run linter

.PHONY: ${MODS_LINT}
${MODS_LINT}: | tools
	cd ./${@F} && ${GOLANGCILINT} run --timeout=15m ./...

.PHONY: test
test: ${MODS_TEST} ## Run tests

.PHONY: ${MODS_TEST}
${MODS_TEST}: | tools
	mkdir -p .output/coverage .output/junit
	cd ./${@F} && ${GOTESTSUM} --junitfile=../.output/junit/${@F}.xml -- -race -covermode=atomic -coverprofile=../.output/coverage/${@F}.txt ./...

.PHONY:tidy ## Tidy all mods
tidy: ${MODS_TIDY} tidy/tools

.PHONY: ${MODS_TIDY}
${MODS_TIDY}:
	cd ./${@F} && go mod tidy

.PHONY: tidy/tools
tidy/tools:
	cd ${TOOLS_DIR} && go mod tidy

.PHONY: download ## Download deps for all mods
download: ${MODS_DOWNLOAD} download/tools

.PHONY: ${MODS_DOWNLOAD}
${MODS_DOWNLOAD}:
	cd ./${@F} && go mod download

.PHONY: download/tools
download/tools:
	cd ${TOOLS_DIR} && go mod download

.PHONY: tidy-check
tidy-check: ${MODS_CHECK} tidy-check/tools ## Check if all mods are tidy
.PHONY: ${MODS_CHECK}
${MODS_CHECK}:
	cd ./${@F} && go mod tidy
	git diff --exit-code --name-status -- ./${@F}/go.mod ./${@F}/go.sum

.PHONY: tidy-check/tools
tidy-check/tools:
	cd ${TOOLS_DIR} && go mod tidy
	git diff --exit-code --name-status -- ${TOOLS_DIR}/go.mod ${TOOLS_DIR}/go.sum

.PHONY: update-deps
update-deps: ${MODS_UPDATE} update-deps/tools ## Update all deps
.PHONY: ${MODS_UPDATE}
GO_VERSION    ?= $(shell go env GOVERSION | sed 's/^go//')

${MODS_UPDATE}:
	cd ./${@F} && go mod edit -go=${GO_VERSION}
	cd ./${@F} && go get $$(go mod edit -json | jq -r '[(.Require[]? | select(.Indirect | not) | .Path)] | map(. + "@latest") | .[]')
	cd ./${@F} && go mod tidy

.PHONY: update-deps/tools
update-deps/tools:
	cd ${TOOLS_DIR} && go mod edit -go=${GO_VERSION}
	cd ${TOOLS_DIR} && go get $$(go mod edit -json | jq -r '[.Tool[]?.Path] | map(. + "@latest") | .[]')
	cd ${TOOLS_DIR} && go mod tidy

.PHONY: api-check
api-check: ${MODS_APICHECK} ## Fail on breaking API changes vs the latest tag

.PHONY: ${MODS_APICHECK}
${MODS_APICHECK}: MOD  = ${@F}
${MODS_APICHECK}: LAST = $(call mod_last_tag,${MOD})
${MODS_APICHECK}: BASE = $(if $(LAST),$(LAST:$(MOD)/%=%),none -version=v1.0.0)
${MODS_APICHECK}: | tools
	cd ${MOD} && ${GORELEASE} -base=${BASE}

.PHONY: tag
tag: ${MODS_TAG} ## Tag any mod that has changes since its last tag

.PHONY: ${MODS_TAG}
${MODS_TAG}: MOD  = ${@F}
${MODS_TAG}: LAST = $(call mod_last_tag,${MOD})
${MODS_TAG}: BASE = $(LAST:$(MOD)/%=%)
${MODS_TAG}: | tools
	@v="v1.0.0"; if [ -n "${LAST}" ]; then \
	  v=$$(cd ${MOD} && ${GORELEASE} -base=${BASE} | tee /dev/stderr | awk '/^Suggested version:/ {print $$3; exit}'); \
	  test -n "$$v" || { echo "${MOD}: gorelease did not suggest a version" >&2; exit 1; }; \
	fi; \
	git tag "${MOD}/$$v" && echo "tagged ${MOD}/$$v"

.PHONY: cross-check
cross-check: ${MODS_CROSS} ## Cross-compile each mod for windows and darwin

.PHONY: ${MODS_CROSS}
${MODS_CROSS}:
	cd ./${@F} && GOOS=windows go build ./...
	cd ./${@F} && GOOS=darwin  go build ./...

.PHONY: clean
clean: ## Clean files
	git clean -Xdf
