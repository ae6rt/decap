## History

* Mark: How hard can it be to build a build server?
* Josh: And why not use a pure Kubernetes backend?

## Overview

A headless CI build server based on a Kubernetes backend that
executes shell script based builds in your build container.

This project is under active development, and has no releases yet.

## Theory of Operation

You have projects you want to build.  Builds are articulated in
terms of shell scripts.  Decap ships with a _base build container_
that can locate those build scripts and run them for you.  

Post commit hooks on your projects of interest drive events into a
web container in the core Decap containerized application which
in turn makes calls into the Kubernetes API master to launch a build
container to build your code.  Builds can also be launched via web
UI.

The base build container locates your build scripts based on 

* the git repository the build scripts are located in
* the subdirectory of that repository where your project resides
* and the branch your scripts should build in your project repository

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

### Sidecar containers

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

See http://aws.amazon.com/documentation/cli/

### IAM user

Decap stores build information in S3 buckets and a DynamoDb
table.  These buckets and table are secured using AWS access policies
that are associated with a dedicated IAM user named _decap_.

In ./aws-resources we provide Decap scripts for creating all the
AWS resources Decap needs.  To run these scripts effectively, you
will need an AWS account with what we're calling _root like_ powers.
That is, an account that can create AWS IAM users, buckets, DynamoDb
tables, and policies.  Your main AWS Dashboard account should have
these powers.

Put your AWS Dashboard account Access Key ID and Secret Access Key
in your $HOME/.aws/credentials file (see [AWS Command Line Client Configuration](http://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html#cli-config-files))

Configure your installation credentials

```
$ cat $HOME/.aws/credentials
[decapadmin]
aws_access_key_id = thekey
aws_secret_access_key = thesecret
```

Configure the default AWS region

```
$ cat $HOME/.aws/config
[decapadmin]
region=us-west-1
```

Create the AWS resources

```
$ sh create-world.sh
```
			
You should now be able to view the following resources in your AWS
Dashboard UI:

* an IAM user named decap
* two S3 buckets, one named decap-console-logs and another named decap-build-artifacts
* one DynamoDb table named decap-build-metadata
* five policies attached to the user decap

The five policies are named:

* decap-db-base
* decap-db-isBuilding
* decap-db-projectKey
* decap-s3-build-artifacts
* decap-s3-console-logs

The create-world.sh script also creates an AWS access key for the
user decap.  The key and secret are written to the file aws.credentials.
Using this access key and secret, the script also creates a Kubernetes Secret
for use in the "Decap Kubernetes Secret for AWS credentials" section below.

## Kubernetes Cluster Setup

Bring up a Kubernetes cluster as appropriate:
https://github.com/kubernetes/kubernetes/tree/master/docs/getting-started-guides

### Decap Kubernetes Secret for AWS and Github credentials 

The Access Key and Secret for user decap created above allow the
Decap webapp to upload build artifacts and console logs to S3 and
make and query entries in the DynmamoDb table.  

However, the webapp also needs access to an OAuth2 Github ClientID
and ClientSecret if you want the webapp to query for branches on a
project's repository for Github-based projects.  See *Project
metadata file* below.  The process for generating a Github OAuth2
credentials for your installation of decap starts here:
https://github.com/settings/applications/new.

The AWS access key and secret will be mounted in the build container
using a [Kubernetes Secret Volume
Mount](https://github.com/kubernetes/kubernetes/blob/master/docs/design/secrets.md).

Craft a decap-system-secrets.yaml augmented with the Github OAuth2
credentials using this as an example:

```
apiVersion: v1
data:
  aws-key: thekey
  aws-secret: base64(thesecret)
  aws-region: base64(theregion)
  github-client-id: base64(ghcid)
  github-client-secret: base64(ghsekrit)
kind: Secret
metadata:
     name: decap-credentials
     namespace: "decap-system"
type: Opaque
```

and create it on the Kubernetes cluster:

```
$ kubectl create -f decap-system-secrets.yaml
```

The base build container will automatically have this Kubernetes
Secret mounted in its container, where the container ENTRYPOINT can
use them for publishing build results.

### Decap Kubernetes Pod creation

Create the pod that runs Decap in the cluster:

```
$ kubectl create -f k8s-resources/decap.yaml
```

## Setting up a build scripts repository

Decap leverages Kubernetes ability to mount a Git repository
readonly inside a container.  When you launch a build in the build
container, Kubernetes will mount the build scripts repo that contains
the build scripts for your projects.  Here is a sample build script
repository

```
https://github.com/ae6rt/decap-build-scripts
```

The build container refers to this repository as a mounted volume
https://github.com/ae6rt/decap/blob/master/web/pod.go#L56.  Build
scripts are indexed by _project key_ by the build container entrypoint
https://github.com/ae6rt/decap/blob/master/build-container/build.sh#L44

```
sh /home/decap/buildscripts/decap-build-scripts/${PROJECT_KEY}/build.sh 2>&1 | tee $CONSOLE
```

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

If this file exists, decap can query the repository manager for
branches on the project.  Knowing the branches, decap can offer to
let the user build a particular branch on project. Github is currently
the only supported repository manager, but Stash and Bitbucket
manager are planned.

### Handling updates to the buildscripts repository

Decap will refresh its representation of the buildscripts repository if you add a post-commit hook to this repository.  Point the handler
at _baseURL_/hooks/buildscripts.

## Base Build Container Environment

Here is the base build container reference:  https://github.com/ae6rt/decap/tree/master/build-container

The following environment variables are available in your build scripts:

* BUILD_ID:  UUID that uniquely identifies this build
* PROJECT_KEY: a composite key consisting of your project parent + library
* BRANCH_TO_BUILD: an optional git branch to build within your application project. This is typically used with Github post commit hook events.

Concurrent builds of a given project + branch are currently forbidden,
and enforced with a lock in etcd, which also runs in the Decap
cluster.

Build pod instances are labelled as follows

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


