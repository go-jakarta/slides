#!/bin/bash

set -ex

#protoc --go_out=plugins=grpc:. hello.proto

protoc \
  -I/usr/local/include \
  -I$GOPATH/src \
  -I$GOPATH/src/github.com/gengo/grpc-gateway/third_party/googleapis \
  -I. \
  --go_out=Mgoogle/api/annotations.proto=github.com/gengo/grpc-gateway/third_party/googleapis/google/api,plugins=grpc:. \
  *.proto
