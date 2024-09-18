#!/usr/bin/env sh

rm -f "$GOBIN/kl"
_THISDIR=$(dirname "$(readlink -f "$0")")
echo "building $(date +"%T")"
if [ "$#" -eq 1 ]; then
  go build -ldflags "-X github.com/robinovitch61/kl/cmd.Version=$1" -o "$_THISDIR/kl" "$_THISDIR/.."
else
  go build -o "$_THISDIR/kl" "$_THISDIR/.."
fi
if [ -f "$_THISDIR/kl" ]; then
  mv "$_THISDIR/kl" "$GOBIN"
  echo "built"
fi
