#!/bin/sh

set -ux

pod=$(kubectl --no-headers --namespace=decap get pods | awk '{print $1}')

kubectl --namespace=decap logs -f $pod

