.PHONY: build test coverage release install
default: build

build:
	bash -ex build.sh

test:
	@go test -count=1 ./... -coverprofile=coverage.out

coverage: test
	@go tool cover -func=coverage.out
	@go tool cover -html=coverage.out
	@rm coverage.out

release:
	bash -ex build.sh cors

install:
	bash -ex build.sh install