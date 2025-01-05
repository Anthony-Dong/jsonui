.PHONY: build test coverage release
default: build

build:
	bash -ex build.sh

test:
	@go test ./... -coverprofile=coverage.out

coverage:
	@go tool cover -func=coverage.out
	@go tool cover -html=coverage.out
	@rm coverage.out

release:
	bash -ex build.sh cors