#!/usr/bin/env bash
# Portable fixture hashing; no effect on the shipped application.
set -euo pipefail
if command -v sha256sum >/dev/null 2>&1;then
    sha256sum "$1"
else
    shasum -a 256 "$1"
fi | awk '{print $1}'
