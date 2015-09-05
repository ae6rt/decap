#!/bin/sh

set -ux

kubectl create -f k8s-resources/aws-secret.yaml
kubectl create -f k8s-resources/decap.yaml 
