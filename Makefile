TEST            ?= ./...
TESTARGS        ?=
TEST_COUNT      ?= 1
TEST_TIMEOUT    ?= 20m
# Cap concurrent acceptance tests against the single shared controller. Defaults
# to GOMAXPROCS otherwise, which overloads the controller and amplifies
# transient-load flakes. Keep small for stability; raise for faster local runs.
TEST_PARALLELISM ?= 4

.PHONY: default
default: build

.PHONY: build
build:
	go install

.PHONY: testacc
testacc:
	go build ./...
	TF_ACC=1 go test $(TEST) -v -count $(TEST_COUNT) -parallel $(TEST_PARALLELISM) -timeout $(TEST_TIMEOUT) $(TESTARGS)
