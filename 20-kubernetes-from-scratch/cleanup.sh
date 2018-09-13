#!/bin/bash

set -ex

HOSTS=$(doctl compute droplet list --format Name,ID|egrep '^k8s.*\.sgp1.gophers.id'|awk '{print $2}')

for i in $HOSTS; do
  doctl compute droplet delete -f $i
done

RECORDS=$(doctl compute domain records list gophers.id --format Name,ID|egrep '^k8s.*\.sgp1'|awk '{print $2}')

for i in $RECORDS; do
  doctl compute domain records delete gophers.id -f $i
done
