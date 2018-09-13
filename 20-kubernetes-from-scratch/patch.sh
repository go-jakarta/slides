#!/bin/bash

kubectl patch service -n kube-system kubernetes-dashboard -p '{"spec":{"externalIPs":["172.18.133.161"]}}'
