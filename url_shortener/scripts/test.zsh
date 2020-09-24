#!/usr/bin/zsh

SCRIPT_DIR=${0:a:h}

. "$SCRIPT_DIR"/vars.zsh

pushd "$ROOT_DIR"

go test -v ./...

popd
