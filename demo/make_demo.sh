#!/usr/bin/env sh

if [ "$(kubectl --context k3d-test get po | wc -l)" -le 3 ]; then
    echo "Error: doesn't seem like you are running the demo k3d cluster"
    exit 1
fi

_THISDIR=$(dirname "$(readlink -f "$0")")

# demo.tape output depends on the current working directory
cd "$_THISDIR" || exit

if [ "$(kl -v)" != "kl demo" ]; then
  "$_THISDIR"/../dev/rebuild.sh demo
fi

vhs "$_THISDIR"/demo.tape && open "$_THISDIR"/demo.gif

cd - || return
