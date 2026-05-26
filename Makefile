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

REMOTE        ?= origin
mod_last_tag   = $(shell git ls-remote --tags --refs ${REMOTE} '$(1)/v*' | sed 's|.*refs/tags/||' | sort -V | tail -1)

.PHONY: all
all: clean tidy-check .WAIT lint test

.NOTPARALLEL: lint
.PHONY: lint
lint: ${MODS_LINT} ## Run linter

.PHONY: ${MODS_LINT}
${MODS_LINT}:
	cd ./${@F} && go tool github.com/golangci/golangci-lint/v2/cmd/golangci-lint run --timeout=15m ./...

.PHONY: test
test: ${MODS_TEST} ## Run tests

.PHONY: ${MODS_TEST}
${MODS_TEST}:
	mkdir -p .output/coverage .output/junit
	cd ./${@F} && go tool gotest.tools/gotestsum --junitfile=../.output/junit/${@F}.xml -- -race -covermode=atomic -coverprofile=../.output/coverage/${@F}.txt ./...

.PHONY:tidy ## Tidy all mods
tidy: ${MODS_TIDY}

.PHONY: ${MODS_TIDY}
${MODS_TIDY}:
	cd ./${@F} && go mod tidy

.PHONY: download ## Download deps for all mods
download: ${MODS_DOWNLOAD}

.PHONY: ${MODS_DOWNLOAD}
${MODS_DOWNLOAD}:
	cd ./${@F} && go mod download

.PHONY: tidy-check
tidy-check: ${MODS_CHECK} ## Check if all mods are tidy
.PHONY: ${MODS_CHECK}
${MODS_CHECK}:
	cd ./${@F} && go mod tidy
	git diff --exit-code --name-status -- ./${@F}/go.mod ./${@F}/go.sum

.PHONY: update-deps
update-deps: ${MODS_UPDATE} ## Update all deps
.PHONY: ${MODS_UPDATE}
GO_VERSION    ?= $(shell go env GOVERSION | sed 's/^go//')

${MODS_UPDATE}:
	cd ./${@F} && go mod edit -go=${GO_VERSION}
	cd ./${@F} && pkgs=$$(go mod edit -json | jq -r '[(.Tool[]?.Path), (.Require[]? | select(.Indirect | not) | .Path)] | map(. + "@latest") | .[]'); \
	  [ -z "$$pkgs" ] || go get $$pkgs
	cd ./${@F} && go mod tidy

.PHONY: api-check
api-check: ${MODS_APICHECK} ## Fail on breaking API changes vs the latest tag

.PHONY: ${MODS_APICHECK}
${MODS_APICHECK}: MOD  = ${@F}
${MODS_APICHECK}: LAST = $(call mod_last_tag,${MOD})
${MODS_APICHECK}: BASE = $(if $(LAST),$(LAST:$(MOD)/%=%),none -version=v1.0.0)
${MODS_APICHECK}:
	cd ${MOD} && go tool gorelease -base=${BASE}

.PHONY: tag
tag: ${MODS_TAG} ## Tag any mod that has changes since its last tag

.PHONY: ${MODS_TAG}
${MODS_TAG}: MOD  = ${@F}
${MODS_TAG}: LAST = $(call mod_last_tag,${MOD})
${MODS_TAG}: BASE = $(LAST:$(MOD)/%=%)
${MODS_TAG}:
	@v="v1.0.0"; if [ -n "${LAST}" ]; then \
	  v=$$(cd ${MOD} && go tool gorelease -base=${BASE} | tee /dev/stderr | awk '/^Suggested version:/ {print $$3; exit}'); \
	  test -n "$$v" || { echo "${MOD}: gorelease did not suggest a version" >&2; exit 1; }; \
	fi; \
	git tag "${MOD}/$$v" && echo "tagged ${MOD}/$$v"

.PHONY: clean
clean: ## Clean files
	git clean -Xdf
