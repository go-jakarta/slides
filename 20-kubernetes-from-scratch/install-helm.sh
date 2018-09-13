#!/bin/bash

set -ex

curl https://raw.githubusercontent.com/kubernetes/helm/master/scripts/get | sudo bash

# install helm tiller
helm init

# fix the helm RBAC
kubectl create serviceaccount \
  --namespace kube-system tiller
kubectl create clusterrolebinding tiller-cluster-rule \
  --clusterrole=cluster-admin \
  --serviceaccount=kube-system:tiller
kubectl patch deploy tiller-deploy \
  --namespace kube-system \
  -p '{"spec":{"template":{"spec":{"serviceAccount":"tiller"}}}}'
