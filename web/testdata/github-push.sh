#!/bin/sh

# Simulate a push event from github.  IP address defaults to minikube cluster.

set -ux

SERVER=${SERVER:=192.168.99.100}
PORT=${PORT:=31000}

OWNER=ae6rt
PROJECT=decap-simple-project

cat <<EOF | curl -v -i -X POST -H"X-Github-Event: push" -H"Content-type: application/json" -d @- http://${SERVER}:${PORT}/hooks/github
{
    "ref": "refs/heads/master", 
    "repository": {
        "full_name": "${OWNER}/${PROJECT}", 
        "id": 35129377, 
        "name": "${PROJECT}", 
        "owner": {
            "email": "${OWNER}@users.noreply.github.com", 
            "name": "${OWNER}"
        }
    }
}
EOF
