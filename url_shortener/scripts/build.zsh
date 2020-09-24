#!/usr/bin/zsh

SCRIPT_DIR=${0:a:h}
MODULE_DIR="$SCRIPT_DIR"/../..
BUILD_DIR="build/url_shortener"
CMD_DIR="url_shortener/cmd"

pushd "$MODULE_DIR"

go  build -o    "$BUILD_DIR"/http_server        "$CMD_DIR"/http_server/main.go
go  build -o    "$BUILD_DIR"/end_to_end_tester  "$CMD_DIR"/end_to_end_tester/main.go
go  build -o    "$BUILD_DIR"/httpload           "$CMD_DIR"/httpload/main.go

popd
