#!/bin/bash

protoc \
  -I. \
  -I/usr/local/include \
  -I${GOPATH}/src \
  -I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
  --go_out=Mgoogle/api/annotations.proto=github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis/google/api,plugins=grpc:$GOPATH/src \
  --grpc-gateway_out=logtostderr=true:. \
  --swagger_out=logtostderr=true:. \
  *.proto
