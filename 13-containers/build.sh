#!/bin/bash

PKG=$(expr "$(pwd)" : "$GOPATH/src/\(.*\)")
CMD=webapp

echo "building $PKG => $GOPATH/src/$PKG/$CMD"

if [[ "$(uname)" != "Linux" ]]; then
  set -ex
  docker run \
    -v $GOPATH/src:/go/src \
    -it golang:1.9 \
    go build -o /go/src/$PKG/$CMD $@ $PKG
else
  set -ex
  go build -o $GOPATH/src/$PKG/$CMD $@ $PKG
fi
