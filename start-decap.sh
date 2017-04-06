#!/bin/sh

set -ux

kubectl create -f k8s-resources/decap-namespaces.yaml

for i in decap decap-system; do
	kubectl --namespace=$i create -f k8s-resources/decap-secrets.yaml
done

kubectl create -f k8s-resources/decap.yaml --record
