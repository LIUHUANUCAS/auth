# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=auth

.PHONY: all build run test clean deps tidy help

all: test build

build:
	$(GOBUILD) -o $(BINARY_NAME) -v

run:
	$(GORUN) main.go

test:
	$(GOTEST) -v ./...

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

deps:
	$(GOGET) -v -t -d ./...

tidy:
	$(GOMOD) tidy

help:
	@echo "Make commands:"
	@echo "  build - Build the application"
	@echo "  run   - Run the application"
	@echo "  test  - Run tests"
	@echo "  clean - Clean build artifacts"
	@echo "  deps  - Get dependencies"
	@echo "  tidy  - Tidy go.mod file"
	@echo "  help  - Show this help message"
