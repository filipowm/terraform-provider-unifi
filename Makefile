TEST         ?= ./...
TESTARGS     ?=
TEST_COUNT   ?= 1
TEST_TIMEOUT ?= 20m

.PHONY: default
default: build

.PHONY: build
build:
	go install

.PHONY: testacc
testacc:
	go build ./...
	TF_LOG_PROVIDER=debug TF_ACC=1 go test $(TEST) -test.parallel 2 -v -count $(TEST_COUNT) -timeout $(TEST_TIMEOUT) $(TESTARGS)
