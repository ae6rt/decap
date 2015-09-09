#!/bin/sh

set -ux

kubectl create -f k8s-resources/decap-namespaces.yaml
kubectl --namespace=decap create -f k8s-resources/decap-secrets.yaml
kubectl --namespace=decap-system create -f k8s-resources/decap-secrets.yaml
kubectl create -f k8s-resources/decap.yaml 
