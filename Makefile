INFO_COLOR=\033[1;34m
RESET=\033[0m
BOLD=\033[1m
TEST ?= $(shell $(GO) list ./... | grep -v vendor)
REVISION = $(shell git describe --always)
ifeq ("$(shell uname)","Darwin")
GO ?= GO111MODULE=on go
else
GO ?= GO111MODULE=on /usr/local/go/bin/go
endif

default: build
ci: depsdev test integration vet lint

depsdev: ## Installing dependencies for development
	$(GO) get golang.org/x/lint/golint

server:
	$(GO) run main.go

build: ## Build as linux binary
	@echo "$(INFO_COLOR)==> $(RESET)$(BOLD)Building$(RESET)"
	$(GO) build -o sshr_bin main.go

test: ## Run test
	@echo "$(INFO_COLOR)==> $(RESET)$(BOLD)Testing$(RESET)"
	$(GO) test -v $(TEST) -timeout=30s -parallel=4
	$(GO) test -race $(TEST)

integration: ## Run integration test after Server wakeup
	@echo "$(INFO_COLOR)==> $(RESET)$(BOLD)Integration Testing$(RESET)"
	$(GO) test -integration -v $(TEST) -timeout=30s -parallel=4

vet: ## Exec go vet
	@echo "$(INFO_COLOR)==> $(RESET)$(BOLD)Vetting$(RESET)"
	$(GO) vet $(TEST)

lint: ## Exec golint
	@echo "$(INFO_COLOR)==> $(RESET)$(BOLD)Linting$(RESET)"
	golint -min_confidence 1.1 -set_exit_status $(TEST)

.PHONY: default test
