#!/usr/bin/zsh

SCRIPT_DIR=${0:a:h}

ROOT_DIR="$SCRIPT_DIR"/../..

BUILD_DIR="build"                           # relative
CMD_DIR="url_shortener/cmd"                 # relative
SHORTENER_DIR="$BUILD_DIR"/url_shortener    # relative

export ROOT_DIR
export BUILD_DIR
export CMD_DIR
export SHORTENER_DIR
