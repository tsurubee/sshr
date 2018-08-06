INFO_COLOR=\033[1;34m
RESET=\033[0m
BOLD=\033[1m

default: server

deps: ## Install dependencies
	@echo "$(INFO_COLOR)==> $(RESET)$(BOLD)Installing Dependencies$(RESET)"
	go get -u github.com/golang/dep/...
	dep ensure

server: deps
	go run main.go

.PHONY: deps depsdev