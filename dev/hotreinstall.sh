#!/usr/bin/env sh

_THISDIR=$(dirname "$(readlink -f "$0")")
find "$_THISDIR/../" -name "*.go" | entr -c "$_THISDIR/reinstall.sh" "$1"
