#!/bin/bash

set -v

docker build . \
  --tag sample-api:latest

docker run \
  --name sample-api \
  --detach \
  --network testnet \
  --rm \
  --publish 8080:8080 \
  sample-api
