#!/bin/bash

MASTER_IP=
MASTER=$MASTER_IP:6443
TOKEN=
HASH=

echo "$MASTER_IP master.k8s.sgp1.gophers.id" | sudo tee -a /etc/hosts

sudo kubeadm join $MASTER \
  --token $TOKEN \
  --discover-token-ca-cert-hash $HASH

