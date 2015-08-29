k8s resources
------------

These are Kubernetes resources to define the aftomato service.

Installation into the cluster
-----------------------------

kubectl create -f aws-secret.yaml

kubectl create -f aftomato.yaml

Launching a build
---------

Builds are currently launched by posting a post-receive hook to the
web service on port 9090.  See web/ in the top level.


