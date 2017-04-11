#!/bin/sh 

set -u

kubectl --namespace=decap-system logs -f $(kubectl --no-headers --namespace=decap-system get pods | awk '{print $1}')" 
