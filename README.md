## History

> Mark: How hard can it be to build a build server?

> Josh: How hard can space travel be?  But whatever.  So why not use a pure Kubernetes backend?

## Overview

Decap is a CI build server based on a Kubernetes backend that
executes shell script-based builds in a specially provided build
pod.  The backend is a containerized webapp that manages build pods,
and the frontend a single page app that provides a friendly UX.
Builds are executed in pods spun up on demand, with build results
published to S3 buckets and a DynamoDb table.

This project is under active development, and has no releases yet.

## Theory of Operation

You have projects you want to build.  Your builds are articulated
in terms of userland shell scripts.  Decap ships with a _base build
container_ that mounts your build scripts as a git repository and
locates them by a _parent/libary_ convention in the container
filesystem.

Either user-initiated builds or post commit hooks sent from your
projects of interest drive HTTP requests to the containerized Decap
webapp.  This webapp in turn makes calls to the Kubernetes API
master to launch an ephemeral build pod to build a single instance
of your code.  Once the build is finished the pod exits, saving no
build pod state from one build to the next.  Build results are
shipped to Amazon AWS S3 buckets and a DynamoDb table.

Your build scripts are completely free form.  Here are two examples:

```
#!/bin/bash

echo hello world
```

```
#!/bin/bash

git clone https://git.example.com/repo --branch ${BRANCH_TO_BUILD}
mvn clean install
mvn deploy
```

### Sidecar build containers

If your build needs additional services, such as MySQL, RabbitMQ,
etc., we plan to provide a way to ingest a Kubernetes Pod descriptor
into the build instruction that will allow you to run these supporting
sidecar services in the build pod along with the main build container.
This means those services will be available to your build at
localhost:port.

## AWS Setup

Decap uses S3 buckets to store build artifacts and console logs,
and DynamoDb to store overall build results and metadata.

### Install the AWS CLI app

To run the scripts below that create AWS resources Decap needs, you
will need to install the AWS Command Line Client.  For instructions
on how to do this, see http://aws.amazon.com/documentation/cli/.

### IAM user

In ./aws-resources we provide shell scripts for creating the AWS
resources Decap needs.  To run these scripts effectively, you will
need an AWS account with what we're calling _root like_ powers.
This account must be capable of creating the following types of
resources on AWS:

* AWS IAM users
* Access credentials
* S3 buckets
* DynamoDb tables
* Policies 

Your main AWS _Dashboard_ account should have these powers.  If it
does not, contact your AWS administrator.

Add your AWS Dashboard account Access Key ID and Secret Access Key
to $HOME/.aws/credentials file (see [AWS Command Line Client
Configuration](http://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html#cli-config-files)),
and a default region to $HOME/.aws/config in a profile named _decapadmin_. 

$HOME/.aws/credentials:

```
[decapadmin]
aws_access_key_id = (your Dashboard key)
aws_secret_access_key = (your Dashboard secret)
```

$HOME/.aws/config:

```
[decapadmin]
region=us-west-1
```

Create the AWS resources Decap requires

```
$ cd aws-resources
$ sh create-world.sh
```

You should now be able to view the following resources in your AWS
Dashboard UI:

* an IAM user named decap
* a set of access credentials for use by the newly created decap user, written to the file aws.credentials
* two S3 buckets: decap-console-logs and decap-build-artifacts
* one DynamoDb table named decap-build-metadata
* five policies attached to the user decap, which are required for access to the S3 buckets and DynamoDb table

The five policies are named:

* decap-db-base
* decap-db-isBuilding
* decap-db-projectKey
* decap-s3-build-artifacts
* decap-s3-console-logs

## Kubernetes Cluster Setup

Bring up a Kubernetes cluster as appropriate:
https://github.com/kubernetes/kubernetes/tree/master/docs/getting-started-guides.
Decap requires SkyDNS, which is included if you use the standard
kube-up.sh script to create your cluster.


## Namespaces

Decap requires two Kubernetes namespaces:  decap and decap-system.
The decap namespace is where your build pod runs.  decap-system is
where the containerized decap webapp runs.

Create the Kubernetes namespaces required for decap

```
$ kubectl create -f k8s-resources/decap-namespaces.yaml
```

### Decap Kubernetes Secret for AWS and Github credentials

The AWS Access Key and Secret for user decap created above allows
the build pod to upload build artifacts and console logs to S3, and
to write build information to the DynamoDb table.  The Access Key
and Secret also allows the decap webapp to access these same buckets
and table.

To be most effective, Decap also needs access to the list of branches
for your various projects.  Decap can query your project repositories
for this branch information.  Without access to your project branch
information, Decap's web UI cannot offer to easily build a particular
branch on your projects.  For Github projects, this means Decap
needs an OAuth2 Github ClientID and ClientSecret.  See *Project
metadata file* below.  Generate Github OAuth2 credentials here:
https://github.com/settings/applications/new.

Using the AWS Access Key and Secret in ./aws-resources/aws.credentials,
and your Github ClientID and ClientSecret, craft a
k8s-resources/decap-secrets.yaml

```
apiVersion: v1
data:
  aws-key: thekey
  aws-secret: base64(thesecret)
  aws-region: base64(theregion)
  github-client-id: base64(github client-id)
  github-client-secret: base64(github client-secret)
kind: Secret
metadata:
     name: decap-credentials
type: Opaque
```

and create it on the Kubernetes cluster in both the _decap_ and
_decap-system_ namespaces

```
$ kubectl --namespace=decap-system create -f k8s-resources/decap-secrets.yaml
$ kubectl --namespace=decap create -f k8s-resources/decap-secrets.yaml
```

The base build container will automatically have these Kubernetes
Secrets mounted in both the build container and the webapp container.
The base build container will use them for publishing build results.
The webapp container will use them for querying AWS and Github.

### Decap Kubernetes Pod creation

Create the pod that runs the Decap webapp in the cluster:

```
$ kubectl create -f k8s-resources/decap.yaml
```

## Setting up a build scripts repository

Decap leverages Kubernetes's ability to mount a Git repository
readonly inside a container.  When you launch a build in the build
container, Kubernetes will mount the build scripts repo that contains
the build scripts for your projects.  Here is a sample build script
repository

```
https://github.com/ae6rt/decap-build-scripts
```

The build container refers to this repository as a mounted volume.
Build scripts are indexed by _project key_ by the build container
entrypoint.  For github based projects, the project key is the
github username + "/" + repository name.  Generally, the username
is referred to as the _parent_ and the repository basename as the
_library_  For example, if the github username is ae6rt and the
repository name is dynamodb-lab, then the project key is
"ae6rt/dynamodb-lab".  The build script is by convention named
build.sh and is located relative to the top level of the build
scipts repository at ae6rt/dynamodb-lab/build.sh.

The build container will call your project's build script, capture
the console logs, and ship the build artifacts, console logs and
build metadata to S3 and DynamoDb.

## Project metadata file

An optional _project.json_ file may be placed on par with a project's
build.sh script.  project.json has the following example format

```
{
     "repo-manager": "github",
     "repo-url": "https://github.com/ae6rt/dynamodb-lab.git",
     "repo-description": "AWS DynamoDb lab"
}
```

If this file exists, Decap will query the repository manager for
branches on the project.  Knowing the branches, the Decap web UI
can offer to let the user build a particular branch on a project.
Github is currently the only supported repository manager, but Stash
and Bitbucket managers are planned.

If the project.json file is absent, Decap will lack information
about your project that allows it to query your project repository
for branch information.

### Handling updates to the build scripts repository

Decap will refresh its representation of the build scripts repository
if you add a post-commit hook to the build scripts repository.
Point the post commit URL at Decap _baseURL_/hooks/buildscripts.  Any
HTTP POST to this endpoint will force a refresh of the build script
repository in the Decap webapp.

## Base Build Container Environment

Here is the base build container reference:
https://github.com/ae6rt/decap/tree/master/build-container

The following environment variables are available in your build scripts:

* BUILD_ID:  UUID that uniquely identifies this build
* PROJECT_KEY: a composite key consisting of your project _parent/library_
* BRANCH_TO_BUILD: an optional git branch for use with builds that can put it to use

Concurrent builds of a given parent/library + branch are currently
forbidden, and enforced with a lock in etcd, which runs in the same
pod as the Decap webapp.

Build pod instances are given the following Kubernetes labels

```
"labels": {
   "type": "decap-build",
   "parent": "{{.Parent}}",
   "library": "{{.Library}}",
   "branch": "{{.BranchToBuild}}",
}
```

## Developing Decap

The Decap source is divided into three parts:

* Base Build Container in build-container/
* Webapp in web/
* Kubernetes resource configs in k8s-resources/

### Base Build Container

This is the place to modify the base build container ENTRYPOINT script and Dockerfile

### Webapp

This is a Go webapp that receives commit hooks from various repository
managers.  Upon receiving a hook on a managed project, Decap will
launch a container to execute a build on the project and branch.

### Kubernetes resource configs

This contains yaml files that describe Kubernetes resources Decap needs to function.
