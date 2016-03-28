#!/bin/bash
protoc --go_out=plugins=grpc:. hello.proto
