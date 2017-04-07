#!/bin/sh

# Simulate a push event from github.  This assumes decap listens on a nodePort 192.168.99.100:31000.  The IP address reflects a minikube cluster.

cat <<EOF | curl -i -X POST -H"X-Github-Event: push" -H"Content-type: application/json" -d @- http://192.168.99.100:31000/hooks/github
{
    "ref": "refs/heads/master", 
    "repository": {
        "full_name": "ae6rt/dynamodb-lab", 
        "id": 35129377, 
        "name": "dynamodb-lab", 
        "owner": {
            "email": "ae6rt@users.noreply.github.com", 
            "name": "ae6rt"
        }
    }
}
EOF
