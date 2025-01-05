#!/usr/bin/env bash

build_flag=("-v" "-ldflags" "-s -w")

# cross_go_build windows amd64
function cross_go_build(){
  binary="jsonui"
  if [ "$1" == "windows" ]; then
    binary="jsonui.exe"
  fi
  CGO_ENABLED=0 GOOS=$1 GOARCH=$2 go build "${build_flag[@]}" -o "bin/$1_$2/$binary"
}

function format_golang_file () {
  project_dir=$(realpath "$1")
	# shellcheck disable=SC2044
	for elem in $(find "${project_dir}" -name '*.go' | grep -v 'example/'); do
		gofmt -w "${elem}" 2>&1;
		goimports -w -srcdir "${project_dir}" -local "$2" "${elem}" 2>&1;
	done
}

case $1 in
cors)
  cross_go_build windows amd64
  cross_go_build darwin amd64
  cross_go_build linux amd64
  cross_go_build linux arm64
  cross_go_build darwin arm64
  zip -j bin/jsonui.zip bin/*
  ;;
format)
  format_golang_file . "github.com/anthony-dong/jsonui"
  ;;
*)
  go build -v -o bin/jsonui .
esac