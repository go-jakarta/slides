#!/bin/bash

set -ex

REGION=sgp1
DOMAIN=gophers.id

# enable external access for 6443
sudo ufw allow 6443/tcp

# enable port for canal
sudo ufw allow 8472/udp

# initialize cluster
sudo kubeadm init \
  --apiserver-advertise-address $(curl -s http://169.254.169.254/metadata/v1/interfaces/private/0/ipv4/address) \
  --pod-network-cidr 10.244.0.0/16 \
  --apiserver-cert-extra-sans k8s-master.$REGION.$DOMAIN \
  --service-dns-domain $DOMAIN

mkdir -p ~/.kube
sudo cp -i /etc/kubernetes/admin.conf ~/.kube/config
sudo chown $(id -u):$(id -g) ~/.kube/config

# setup networking
mkdir canal

pushd canal &> /dev/null

wget https://docs.projectcalico.org/v3.2/getting-started/kubernetes/installation/hosted/canal/rbac.yaml
wget https://docs.projectcalico.org/v3.2/getting-started/kubernetes/installation/hosted/canal/canal.yaml
perl -pi -e 's/canal_iface:.*/canal_iface: "eth1"/' canal.yaml

# apply canal config
kubectl apply -f rbac.yaml
kubectl apply -f canal.yaml

# taint master node
kubectl taint nodes --all node-role.kubernetes.io/master-

popd &> /dev/null
