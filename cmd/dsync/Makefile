GO := go
GO_BUILD := $(GO) build
GO_CLEAN := $(GO) clean

TEST_SCRIPT := ./test/main.sh

all: build

build:
	$(GO_BUILD) -o my_go_project main.go

clean:
	$(GO_CLEAN)

test:
	$(TEST_SCRIPT)

run:
	./my_go_project

.PHONY: all build clean test run