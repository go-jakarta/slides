#!/bin/bash

set -ex

REGION=sgp1
DOMAIN=gophers.id
SIZE=s-4vcpu-8gb
IMAGE=ubuntu-16-04-x64

for i in $(seq 1 3); do
  HOST=k8s0$i
  NAME="$HOST.$REGION.$DOMAIN"
  doctl compute droplet create $NAME \
    --region $REGION \
    --size $SIZE \
    --image $IMAGE \
    --enable-private-networking \
    --user-data-file cloud-config.yaml \
    --wait

  IP=$(doctl compute droplet list --format Name,PublicIPv4|grep ^$NAME|awk '{print $2}')

  # cleanup existing dns records
  RECORDS=$(doctl compute domain records list $DOMAIN --format Name,ID|egrep "^$HOST.$REGION"|awk '{print $2}')
  for rec in $RECORDS; do
    doctl compute domain records delete $DOMAIN -f $rec
  done

  # create dns record
  doctl compute domain records create $DOMAIN \
    --record-name $HOST.$REGION \
    --record-data $IP \
    --record-type A \
    --record-ttl 60

  if [[ "$i" == "1" ]]; then
    doctl compute domain records create $DOMAIN \
      --record-name k8s-master.$REGION \
      --record-data $IP \
      --record-type A \
      --record-ttl 60
  fi
done
