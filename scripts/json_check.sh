#!/usr/bin/env bash
# Source from maintainer tooling before consuming a single-object record.
json_object() {
    [[ -f "$1" && ! -L "$1" ]] && jq -e -s 'length==1 and (.[0]|type=="object")' "$1" >/dev/null
}
