SHELL := /usr/bin/env bash

# Disable built-in rules
MAKEFLAGS += --no-builtin-rules
MAKEFLAGS += --no-builtin-variables
.SUFFIXES:
.SECONDARY:

.DEFAULT_GOAL := help
.PHONY: help
help: ## Show this help
	@grep -E -h '\s##\s' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = "(: ).*?## "}; {gsub(/\\:/,":",$$1)}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: fmt
fmt: ## Run 'go fmt' against code
	go fmt ./... ./examples/

.PHONY: vet
vet: ## Run 'go vet' against code
	go vet -tags=examples ./...

.PHONY: lint
lint: fmt vet ## Invokes the fmt and vet targets
	@echo 'Check for uncommitted changes ...'
	git diff --exit-code

.PHONY: clean
clean: ## Clean the project
	@rm -rf cover.out

.PHONY: test
test: ## Run unit tests
	@go test -race -coverprofile cover.out -covermode atomic -count 1 -tags=examples ./...
