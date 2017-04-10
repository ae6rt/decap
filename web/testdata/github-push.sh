#!/bin/sh

# Simulate a push event from github.  This assumes decap listens on a nodePort 192.168.99.100:31000.  The IP address reflects a minikube cluster.

set -ux

readonly OWNER=ae6rt
readonly PROJECT=decap-simple-project

cat <<EOF | curl -v -i -X POST -H"X-Github-Event: push" -H"Content-type: application/json" -d @- http://192.168.99.100:31000/hooks/github
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
