VERSION := $(shell git describe --tags)
BUILD := $(shell git rev-parse --short HEAD)
PROJECTNAME := $(shell basename "$(PWD)")

install:
	go get -u -v -f all

## test: Run all unit-test
test:
	@echo "  >  building binary..."
	go test -v -race -timeout 60s ./...




all: help
help: Makefile
	@echo
	@echo " Choose a command run in "$(PROJECTNAME)":"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo


.PHONY: all clean install uninstall test build compile help
.DEFAULT_GOAL := help

