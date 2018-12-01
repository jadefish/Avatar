PROJECT := avatar
PKG := github.com/jadefish
BINARIES := login

# Go and tools:
GO ?= $(shell command -v go)
FMT ?= $(GO) fmt
VET ?= $(GO) vet

# Platform, architecture, and build flags:
GOOS ?= linux
GOARCH ?= amd64
GOFLAGS ?= -v

# Directories and files:
SRC_FILES = $(wildcard **/*.go)
BIN_DIR ?= bin
PKG := $(BINARIES:%=$(PKG)/$(PROJECT)/cmd/%)
TARGETS := $(BINARIES:%=$(BIN_DIR)/%)

# Pretty.
GREEN := $(shell printf "\033[32m")
BLUE := $(shell printf "\033[34m")
RESET := $(shell printf "\033[0m")

.PHONY: go fmt vet deploy clean
build: $(TARGETS)
default: $(TARGETS) fmt vet

$(TARGETS): go $(BIN_DIR) $(SRC_FILES)
	$(info $(BLUE)Building$(RESET) $@ $(BLUE)for$(RESET) $(GOARCH)/$(GOOS)...)
	env GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build $(GOFLAGS) -o $@ $(PKG)

$(BIN_DIR):
	test -d $(BIN_DIR) || mkdir -p $(BIN_DIR)

fmt:
	$(FMT) ./...

vet:
	$(VET) ./...

clean:
	test -d $(BIN_DIR) && $(RM) -r $(BIN_DIR)

go:
ifdef GO
	$(info Found $(GREEN)go$(RESET): $(GO))
else
	$(error Unable to locate go)
endif
