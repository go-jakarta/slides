#!/bin/bash

set -v

docker build . \
  --tag sample-aks:latest

docker run \
  --name sample-aks \
  --detach \
  --network testnet \
  --rm \
  --publish 8080:8080 \
  sample-aks
