PROJECT := avatar
PKG := github.com/jadefish
BINARIES := login

# Go and dep:
GO ?= $(shell command -v go)
DEP ?= $(shell command -v dep)

# Platform, architecture, and build flags:
GOOS ?= linux
GOARCH ?= amd64
GOFLAGS ?= -v

# Directories and files:
SRC_FILES = $(wildcard **/*.go)
BIN_DIR ?= bin
VENDOR_DIR ?= vendor
PKG := $(BINARIES:%=$(PKG)/$(PROJECT)/cmd/%)
TARGETS := $(BINARIES:%=$(BIN_DIR)/%)

# Pretty.
GREEN := $(shell printf "\033[32m")
BLUE := $(shell printf "\033[34m")
RESET := $(shell printf "\033[0m")

.PHONY: dep go fmt vet deploy clean
build: $(TARGETS)
default: $(TARGETS)

$(TARGETS): go $(BIN_DIR) $(SRC_FILES) $(VENDOR_DIR) fmt vet
	$(info $(BLUE)Building$(RESET) $@ $(BLUE)for$(RESET) $(GOOS)/$(GOARCH)...)
	env GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build $(GOFLAGS) -o $@ $(PKG)

$(BIN_DIR):
	test -d $(BIN_DIR) || mkdir -p $(BIN_DIR)

$(VENDOR_DIR): dep
	$(info $(BLUE)Fetching dependencies...$(RESET))
	@$(DEP) ensure

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

clean:
	test -d $(BIN_DIR) && $(RM) -r $(BIN_DIR)
	test -d $(VENDOR_DIR) && $(RM) -r $(VENDOR_DIR)

go:
ifdef GO
	$(info Found $(GREEN)go$(RESET): $(GO))
else
	$(error Unable to locate go)
endif

dep:
ifdef DEP
	$(info Found $(GREEN)dep$(RESET): $(DEP))
else
	$(error Unable to locate dep)
endif
