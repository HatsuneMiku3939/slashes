all: help

.PHONY : help
help : Makefile
	@sed -n 's/^##//p' $< | awk 'BEGIN {FS = ":"}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

TOOLS_MOD_DIR := ./tools
TOOLS_DIR := $(abspath ./.tools)
$(TOOLS_DIR)/golangci-lint: $(TOOLS_MOD_DIR)/go.mod $(TOOLS_MOD_DIR)/go.sum $(TOOLS_MOD_DIR)/tools.go
	@echo BUILD golangci-lint
	@cd $(TOOLS_MOD_DIR) && \
	go build -o $(TOOLS_DIR)/golangci-lint github.com/golangci/golangci-lint/cmd/golangci-lint

$(TOOLS_DIR)/mockery: $(TOOLS_MOD_DIR)/go.mod $(TOOLS_MOD_DIR)/go.sum $(TOOLS_MOD_DIR)/tools.go
	@echo BUILD mockery
	@cd $(TOOLS_MOD_DIR) && \
	go build -o $(TOOLS_DIR)/mockery github.com/vektra/mockery/v2

$(TOOLS_DIR)/cobra: $(TOOLS_MOD_DIR)/go.mod $(TOOLS_MOD_DIR)/go.sum $(TOOLS_MOD_DIR)/tools.go
	@echo BUILD cobra
	@cd $(TOOLS_MOD_DIR) && \
	go build -o $(TOOLS_DIR)/cobra github.com/spf13/cobra-cli

$(TOOLS_DIR)/godoc: $(TOOLS_MOD_DIR)/go.mod $(TOOLS_MOD_DIR)/go.sum $(TOOLS_MOD_DIR)/tools.go
	@echo BUILD godoc
	@cd $(TOOLS_MOD_DIR) && \
	go build -o $(TOOLS_DIR)/godoc golang.org/x/tools/cmd/godoc

## tools: Build all tools
tools: $(TOOLS_DIR)/mockery $(TOOLS_DIR)/golangci-lint $(TOOLS_DIR)/cobra $(TOOLS_DIR)/godoc

## lint: Run golangci-lint
.PHONY: lint
lint: $(TOOLS_DIR)/golangci-lint gen
	@echo GO LINT
	@$(TOOLS_DIR)/golangci-lint run -c .github/linters/.golangci.yaml --out-format colored-line-number
	@printf "GO LINT... \033[0;32m [OK] \033[0m\n"

## test: Run test
.PHONY: test
test: gen
	@echo TEST
	@go test ./...
	@printf "TEST... \033[0;32m [OK] \033[0m\n"

## test/coverage: Run test and generate coverage report
.PHONY: test/coverage
test/coverage: gen
	@go test  ./... -coverprofile=coverage.txt -covermode=atomic
	@go tool cover -html=coverage.txt -o coverage.html

## gen: Run all code generator
GEN_TARGETS=gen/mock
.PHONY: gen
gen: $(GEN_TARGETS)

## gen/mock: Run mock generator
.PHONY: $(GEN_TARGETS)
gen/mock: $(TOOLS_DIR)/mockery
	@echo GENERATE mocks
	@go generate ./...

## godoc: View godoc
PKG_NAME:=$(shell cat go.mod | grep module | cut -d' ' -f2)
.PHONY: godoc
godoc: $(TOOLS_DIR)/godoc
	@echo "Open http://localhost:6060/pkg/$(PKG_NAME) on browser."
	$(TOOLS_DIR)/godoc -http localhost:6060

## build: Build binary
.PHONY: build
build:
	@echo BUILD
	@go build -o bin/slashes ./cmd/slashes
	@printf "BUILD... \033[0;32m [OK] \033[0m\n"

.PHONY: lint-ci
lint-ci: $(TOOLS_DIR)/golangci-lint
	@echo GO LINT
	@$(TOOLS_DIR)/golangci-lint run -c .github/linters/.golangci.yaml
	@printf "GO LINT... \033[0;32m [OK] \033[0m\n"
