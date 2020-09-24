#!/usr/bin/zsh

SCRIPT_DIR=${0:a:h}

. "$SCRIPT_DIR"/vars.zsh

pushd "$ROOT_DIR"

go  build -o    "$SHORTENER_DIR"/http_server        "$CMD_DIR"/http_server/main.go
go  build -o    "$SHORTENER_DIR"/end_to_end_tester  "$CMD_DIR"/end_to_end_tester/main.go
go  build -o    "$SHORTENER_DIR"/httpload           "$CMD_DIR"/httpload/main.go

popd
