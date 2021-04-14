SHELL := /bin/bash
.DEFAULT_GOAL := help

help: ## Show this help
	@echo Dependencies: go
	@egrep -h '\s##\s' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

test: ## Run library tests
	go test

example: eyecenters.go ## Run example (generates artifact)
	go run example/main.go
