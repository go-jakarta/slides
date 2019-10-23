#!/bin/bash

set -v

docker build . \
  --tag sample-api:latest

docker run \
  --name sample-api \
  --network testnet \
  --rm \
  --publish 8080:8080 \
  sample-api
