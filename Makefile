all: fmt-check lint-check build

BINDIR := bin

LINTER_VERSION := 1.45.2
LINTER := $(BINDIR)/golangci-lint_$(LINTER_VERSION)
DEV_OS := $(shell uname -s | tr A-Z a-z)

$(LINTER):
	mkdir -p $(BINDIR)
	wget "https://github.com/golangci/golangci-lint/releases/download/v$(LINTER_VERSION)/golangci-lint-$(LINTER_VERSION)-$(DEV_OS)-amd64.tar.gz" -O - \
		| tar -xz -C $(BINDIR) --strip-components=1 --exclude=README.md --exclude=LICENSE
	mv $(BINDIR)/golangci-lint $(LINTER)

.PHONY: fmt-check
fmt-check:
	BADFILES=$$(gofmt -l -s -d $$(find . -type f -name '*.go')) && [ -z "$$BADFILES" ] && exit 0

.PHONY: lint-check
lint-check: $(LINTER)
	$(LINTER) run --deadline=2m

## DEBUG BUILDS

GO_FILES = $(shell find . -type f -name '*.go')

$(BINDIR)/gpublame: $(GO_FILES)
	mkdir -p $(BINDIR)
	go build -o $(BINDIR)/gpublame ./cmd

.PHONY: build
build: $(BINDIR)/gpublame

## RELEASE BUILDS

RELEASEDIR := $(BINDIR)/release
ARCHES := amd64 arm arm64

# This rule expects targets with the format $(RELEASEDIR)/gpublame-GOARCH
$(RELEASEDIR)/gpublame-%: $(GO_FILES)
	mkdir -p $(RELEASEDIR)
	GOARCH=$(word 2,$(subst -, ,$@)) \
	  go build -o $@ -ldflags "-s -w" ./cmd

.PHONY: release
release: $(foreach arch,$(ARCHES),$(RELEASEDIR)/gpublame-$(arch))
