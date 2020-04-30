PROJECT := avatar
PKG := github.com/jadefish
BINARIES := login game

# Go and tools:
GO ?= $(shell command -v go)
FMT ?= $(GO) fmt
VET ?= $(GO) vet

# Platform, architecture, and build flags:
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
GOFLAGS ?= -v

# Directories and files:
SRC_FILES = $(wildcard **/*.go)
BIN_DIR ?= bin
COVERAGE_DIR ?= coverage
PKG := $(BINARIES:%=$(PKG)/$(PROJECT)/cmd/%)
TARGETS := $(BINARIES:%=$(BIN_DIR)/%)

.PHONY: go fmt vet deploy clean test coverage
build: $(TARGETS)
default: $(TARGETS) fmt vet

$(TARGETS): go $(BIN_DIR) $(SRC_FILES)
	env GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build $(GOFLAGS) -o $(BIN_DIR) $(PKG)

$(BIN_DIR):
	test -d $(BIN_DIR) || mkdir -p $(BIN_DIR)

fmt:
	$(FMT) ./...

vet:
	$(VET) ./...

clean:
	test -d $(BIN_DIR) && $(RM) -r $(BIN_DIR)
	test -d $(COVERAGE_DIR) && $(RM) -r $(COVERAGE_DIR)

test:
	$(GO) test ./...

coverage:
	$(GO) test -coverprofile=$(COVERAGE_DIR)/cover.out ./...
	$(GO) tool cover -html=$(COVERAGE_DIR)/cover.out -o $(COVERAGE_DIR)/cover.html

go:
ifndef GO
	$(error Unable to locate go)
endif
