#!/usr/bin/env sh

rm -f "$GOBIN/kl"
_THISDIR=$(dirname "$(readlink -f "$0")")
echo "installing $(date +"%T")"
go install "$_THISDIR/.."
echo "installed"
