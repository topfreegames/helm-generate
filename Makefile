# Makefile for Helm Generate
#
# This Makefile makes an effort to provide standard make targets, as describe
# by https://www.gnu.org/prep/standards/html_node/Standard-Targets.html.

# include other Makefiles
include Makefile.*

DOCKER_IMAGE := "helm-template"
PROJECT := "helm-generate"
TAG := $(shell git tag -l --points-at HEAD)
COMMIT := $(shell git describe --always --long --dirty --tags)
VERSION := $(shell [ ! -z "${TAG}" ] && echo "${TAG}" || echo "${COMMIT}")
REVISION := $(shell git rev-parse HEAD)
SHELL := /bin/sh
GITREMOTE := "https://git.topfreegames.com/sre/helm-generate"

GOLANGCI_LINT := golangci-lint

SOURCES := $(shell \
	find . -name '*.go' | \
	grep -Ev './(build|third_party|vendor)/' | \
	xargs)
ifdef DEBUG
$(info SOURCES = $(SOURCES))
endif

# Go packages to compile / test.
PACKAGES ?= ./...
ifdef DEBUG
$(info PACKAGES = $(PACKAGES))
endif

# Make is verbose in Linux. Make it silent.
MAKEFLAGS += --silent

# Use linker flags to provide version settings to the target
# # Also build it with as much as possible static links. It may do the build a bit slower
LDFLAGS=-ldflags "-X=${GITREMOTE}/pkg/constants.Version=$(VERSION) -extldflags '-static'"

################################################################################
## Standard make targets
################################################################################

##- all: Run fix, build and install steps.
.DEFAULT_GOAL := all
.PHONY: all
all: fix build install

##- build: Build go binary
.PHONY: version
version:
	@echo "$(VERSION)"

##- build: Build go binary
.PHONY: build
build: go-build

##- install: Install binary on system
.PHONY: install
install: go-install

##- uninstall: Uninstall binary
.PHONY: uninstall
uninstall: go-uninstall

##- clean: Clean build files. Runs `go clean` internally.
.PHONY: clean
clean: go-clean

##- test: Run all tests
.PHONY: test
test: go-test

##- help: Show options
.PHONY: help
help: Makefile
	@echo
	@echo " Choose a command run in "$(PROJECT)":"
	@echo
	@sed -n 's/^##-//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo

##- build-docker: Builds application in a docker image
.PHONY: build-docker
build-docker:
	@echo " > Building docker image"
	docker build -t $(DOCKER_IMAGE) .

################################################################################
## Go-like targets
################################################################################

go-build:
	@echo "  >  Building binary"
	@mkdir -p $(GOBINDIR)
	$(GOBUILD) -o $(GOBINDIR) $(LDFLAGS) $(PACKAGES)

go-install:
	@echo "  >  Installing ${PROJECT}"
	@GOBINDIR=$(GOBINDIR) $(GOINSTALL) $(PACKAGES)

go-uninstall:
	@echo "  > Uninstalling ${PROJECT}"
	$(RM) $(GOPATH)/bin/$(PROJECT)

go-clean:
	@echo "  >  Cleaning build cache"
	@GOBINDIR=$(GOBINDIR) $(GOCLEAN)

go-test:
	@echo "  >  Executing tests"
	@sh -c "${GOTEST} -cover -coverprofile coverage.out -v -tags=unit ${SILENT_CMD_SUFFIX} ${PACKAGES}"

################################################################################
## Linters and formatters
################################################################################

##- fix: Add and remove missing dependencies
.PHONY: fix
fix:
	@echo "  >  Making sure go.mod matches the source code"
	$(GOMOD) vendor
	$(GOMOD) tidy

##- lint: Run lint
.PHONY: lint
lint:
	@echo "  >  Running lint"
	$(GOLANGCI_LINT) run $(PACKAGES) --timeout=120s
