PROJECT = Avatar
GO ?= $(shell command -v go)
DEP ?= $(shell command -v dep)
DOCKER_COMPOSE ?= $(shell command -v docker-compose)

GOOS ?= linux
GOARCH ?= amd64

SRC_FILES = $(wilcard **/*.go)
BIN_DIR ?= bin
VENDOR_DIR ?= vendor

TARGET ?= $(BIN_DIR)/$(PROJECT)

GREEN := $(shell printf "\033[32m")
BLUE := $(shell printf "\033[34m")
RESET := $(shell printf "\033[0m")

.PHONY: dep go fmt vet deploy clean
build: $(TARGET)
default: $(TARGET)

$(TARGET): go $(BIN_DIR) $(SRC_FILES) fmt vet $(VENDOR_DIR)
	$(info $(BLUE)Building for:$(RESET) $(GOOS)/$(GOARCH))
	env GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build -o $@ -v

$(BIN_DIR):
	test -d $@ || mkdir -p $@

$(VENDOR_DIR): dep
	@$(DEP) ensure
	$(info $(BLUE)Fetching dependencies...$(RESET))

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

deploy: $(TARGET)
	$(DOCKER_COMPOSE) build
	$(DOCKER_COMPOSE) up -d

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

docker:
ifdef DOCKER
	$(info Found $(GREEN)docker$(RESET): $(DOCKER))
else
	$(error Unable to locate docker)
endif
