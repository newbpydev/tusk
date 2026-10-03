#!/usr/bin/env bash
jq_binary_option=''
case "${OSTYPE:-}" in msys*|cygwin*) jq_binary_option=--binary ;; esac
# Source from maintainer tooling before consuming a single-object record.
json_object() {
    [[ -f "$1" && ! -L "$1" ]] && jq ${jq_binary_option:+"--binary"} -e -s 'length==1 and (.[0]|type=="object")' "$1" >/dev/null
}
