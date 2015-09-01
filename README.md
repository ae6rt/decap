## Overview

A CI build server based on a Kubernetes backend that executes builds
in a build container.

This project is under active development, and has no releases yet.

_Aftomato_ is loosely based on the Greek word for _automation_.

## Theory of Operation

You have projects you want to build.  Builds are articulated in
terms of shell scripts.  Aftomato ships with a _base build container_
that can locate those build scripts and run them for you.  

Post commit hooks on your projects of interest drive events into a
web container in the core Aftomato containerized application which
in turn induce calls into the Kubernetes API master to launch a
build container to build your code.  Builds can also be launched
via web UI.

The build container locates your build scripts based on 

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

Aftomato uses S3 buckets to store build artifacts and console logs,
and DynamoDb to store overall build results and metadata.

### Install the AWS CLI app

See http://aws.amazon.com/documentation/cli/

### IAM user

Aftomato stores build information in S3 buckets and a DynamoDb
table.  These buckets and table are secured using AWS access policies
that are associated with a dedicated IAM user named _aftomato_.

Create the IAM user named _aftomato_ in your AWS Console.  This
user needs no SSH private key, but it will need the usual AWS Access
Key and associated Secret.  When you create the user, AWS Console
will display the access key and secret.  Save these to a file ---
the Aftomato build container will need them to publish build results,
and the web frontend to Aftomato will need them to perform build
queries and administration.

View the new user in your AWS Console and note its Amazon Resource
Name (ARN).  It will look like this

```
User ARN: arn:aws:iam::<your account ID>:user/aftomato

```

where _your account ID_ is your Amazon account ID.

### Buckets

Create two S3 buckets, one for build artifacts (even if you don't
think you'll have any) and build console logs.

<table>
    <tr>
        <th>Bucket Name</th>
        <th>Purpose</th>
    </tr>
    <tr>
        <td>aftomato-build-artifacts</td>
        <td>store build artifacts</td>
    </tr>
    <tr>
        <td>aftomato-console-logs</td>
        <td>store build console logs</td>
    </tr>
</table>

#### Bucket policy

The _aftomato_ user created earlier must be given read/write/list
permissions on each S3 bucket in the table above.  Here are example
policies that show correct form, but will not actually work because
we have changed the Id and Statement Ids for security purposes.
Use the [AWS Policy
Generator](http://awspolicygen.s3.amazonaws.com/policygen.html) to
craft your unique policies with unique Ids.


A sample build artifact bucket policy:

```
{
	"Version": "2012-10-17",
	"Id": "<some policy ID"",
	"Statement": [
		{
			"Sid": "<some statement ID>",
			"Effect": "Allow",
			"Principal": {
				"AWS": "arn:aws:iam::<your account ID>:user/aftomato"
			},
			"Action": [
				"s3:DeleteObject",
				"s3:GetObject",
				"s3:PutObject"
			],
			"Resource": "arn:aws:s3:::aftomato-build-artifacts/*"
		},
		{
			"Sid": "<some other statement ID>",
			"Effect": "Allow",
			"Principal": {
				"AWS": "arn:aws:iam::<your account ID>:user/aftomato"
			},
			"Action": "s3:ListBucket",
			"Resource": "arn:aws:s3:::aftomato-build-artifacts"
		}
	]
}
```

A sample console log bucket policy:

```
{
	"Version": "2012-10-17",
	"Id": "<some policy ID"",
	"Statement": [
		{
			"Sid": "<some statement ID>",
			"Effect": "Allow",
			"Principal": {
				"AWS": "arn:aws:iam::<your account ID>:user/aftomato"
			},
			"Action": [
				"s3:DeleteObject",
				"s3:GetObject",
				"s3:PutObject"
			],
			"Resource": "arn:aws:s3:::aftomato-console-logs/*"
		},
		{
			"Sid": "<some other statement ID>",
			"Effect": "Allow",
			"Principal": {
				"AWS": "arn:aws:iam::<your account ID>:user/aftomato"
			},
			"Action": "s3:ListBucket",
			"Resource": "arn:aws:s3:::aftomato-console-logs"
		}
	]
}
```

Select the bucket of interest in the AWS S3 Console area, and attach
these policies to their respective buckets.

### DynamoDb

In your AWS Console create a DynamoDb table named _aftomato-build-metadata_
in your preferred region that has these properties.  The table
should have a main hashkey name _buildID_ with no range key, and a
global secondary index with hashkey _projectKey_ and range _buildTime_.

N.B. You are responsible for getting the _ProvisionedThroughput_
right based on your anticipated usage, which is related to how much
Amazon will bill you for that usage.

This example shows the highlights:

```
$ aws --profile <your AWS credentials profile> dynamodb describe-table --region us-west-1 --table-name aftomato-build-metadata
{
    "Table": {
        "GlobalSecondaryIndexes": [
            {
                "IndexSizeBytes": 0,
                "IndexName": "projectKey-buildTime-index",
                "Projection": {
                    "ProjectionType": "ALL"
                },
                "ProvisionedThroughput": {
                    "NumberOfDecreasesToday": 0,
                    "WriteCapacityUnits": 1,
                    "ReadCapacityUnits": 1
                },
                "IndexStatus": "ACTIVE",
                "KeySchema": [
                    {
                        "KeyType": "HASH",
                        "AttributeName": "projectKey"
                    },
                    {
                        "KeyType": "RANGE",
                        "AttributeName": "buildTime"
                    }
                ],
                "ItemCount": 0
            }
        ],
        "AttributeDefinitions": [
            {
                "AttributeName": "buildID",
                "AttributeType": "S"
            },
            {
                "AttributeName": "buildTime",
                "AttributeType": "N"
            },
            {
                "AttributeName": "projectKey",
                "AttributeType": "S"
            }
        ],
        "ProvisionedThroughput": {
            "NumberOfDecreasesToday": 0,
            "WriteCapacityUnits": 1,
            "ReadCapacityUnits": 1
        },
        "TableSizeBytes": 0,
        "TableName": "aftomato-build-metadata",
        "TableStatus": "ACTIVE",
        "KeySchema": [
            {
                "KeyType": "HASH",
                "AttributeName": "buildID"
            }
        ],
        "ItemCount": 0,
        "CreationDateTime": 1440946400.106
    }
}
```

#### DynamoDb access policy

Crafting the IAM policy for DynamoDb is a bit different from that
of crafting bucket policies.  We first craft a IAM Policy for the
DynamoDb table access, then attach that to the _aftomato_ IAM user.
We choose to split the policies up into three parts, one for the
database r/w operations, and two others for r/w on global secondary
indexes of interest.

The main database policy:
```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "<some statement ID>",
            "Effect": "Allow",
            "Action": [
                "dynamodb:*"
            ],
            "Resource": [
                "arn:aws:dynamodb:us-west-1:<your account ID>:table/aftomato-build-metadata"
            ]
        }
    ]
}
```

The index on projectKey-buildTime:

```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "<some other statement ID>",
            "Effect": "Allow",
            "Action": [
                "dynamodb:*"
            ],
            "Resource": [
                "arn:aws:dynamodb:us-west-1:<your account ID>:table/aftomato-build-metadata/index/projectKey-buildTime-index"
            ]
        }
    ]
}
```

The index on isBuilding:

```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "<yet another statement ID>",
            "Effect": "Allow",
            "Action": [
                "dynamodb:*"
            ],
            "Resource": [
                "arn:aws:dynamodb:us-west-1:<your account ID>:table/aftomato-build-metadata/index/isBuilding-index"
            ]
        }
    ]
}
```

adjusting the AWS region as appropriate to where you created the DynamoDb table.

In the Users section of the AWS Console, attach these policies to the _aftomato_ user.

### Smoke testing AWS buckets and DynamoDb

If you have configured the AWS policies correctly for the Aftomato
buckets and DynamoDb table, you should be able to do things like
this

```
$ aws --profile aftomato s3 cp /etc/hosts s3://aftomato-build-artifacts/hosts.txt
upload: ../../../../../../../../../etc/hosts to s3://aftomato-build-artifacts/hosts.txt

$ aws --profile aftomato s3 cp /etc/hosts s3://aftomato-console-logs/hosts.txt
upload: ../../../../../../../../../etc/hosts to s3://console-logs/hosts.txt

$ aws --profile aftomato  dynamodb describe-table --table-name aftomato-build-metadata
{
    "Table": {
        "GlobalSecondaryIndexes": [
            {
                "IndexSizeBytes": 0,
                "IndexName": "projectKey-buildTime-index",
                "Projection": {
                    "ProjectionType": "ALL"
                },
...
```

where the aftomato AWS credentials are configured in $HOME/.aws/credentials thusly

```
[aftomato]
aws_access_key_id = <your access key>
aws_secret_access_key = <your secret>
region=us-west-1
```

## Kubernetes Cluster Setup

Bring up a Kubernetes cluster as appropriate:
https://github.com/kubernetes/kubernetes/tree/master/docs/getting-started-guides

### Aftomato Kubernetes Secret for AWS credentials

The AWS access key and secret will be mounted in the build container
using a [Kubernetes Secret Volume
Mount](https://github.com/kubernetes/kubernetes/blob/master/docs/design/secrets.md).

As shipped with Aftomato, the Kubernetes Secret looks like this

```
$ cat k8s-resources/aws-secret.yaml
apiVersion: v1
data:
  aws-key: thekey
  aws-secret: thesecret
  region: theregion
kind: Secret
metadata:
     name: aftomato-aws-credentials
type: Opaque

```

_thekey_, _thesecret_, and _theregion_ are the _aftomato_ AWS IAM
User's Access Key, Secret, and default region, respectively.  Replace
these values with their respective Base64 encoded representations

```
$ echo -n "mykey" | openssl base64
bXlrZXk=

```

```
$ echo -n "mysekrit" | openssl base64
bXlzZWtyaXQ=
```

```
$ echo -n "us-west-1" | openssl base64
dXMtd2VzdC0x
```

to produce the production ready Kubernetes Secret

```
$ cat k8s-resources/aws-secret.yaml
apiVersion: v1
data:
  aws-key: bXlrZXk=
  aws-secret: bXlzZWtyaXQ=
  region: dXMtd2VzdC0x
kind: Secret
metadata:
     name: aftomato-aws-credentials
type: Opaque

```

Create this Kubernetes Secret in the Kubernetes cluster with kubectl

```
$ kubectl create -f k8s-resources/aws-secret.yaml
```

The base build container will automatically have this Kubernetes
Secret mounted in its container, where the container ENTRYPOINT can
use them for publishing build results.


### Aftomato Kubernetes Pod creation

Create the pod that runs Aftomato in the cluster:

```
$ kubectl create -f k8s-resources/aftomato.yaml
```

## Setting up a build scripts repository

Aftomato leverages Kubernetes ability to mount a Git repository
readonly inside a container.  When you launch a build in the build
container, Kubernetes will mount the build scripts repo that contains
the build scripts for your projects.  Here is a sample build script
repository

```
https://github.com/ae6rt/aftomato-build-scripts
```

The build container refers to this repository as a mounted volume
https://github.com/ae6rt/aftomato/blob/master/web/pod.go#L56.  Build
scripts are indexed by _project key_ by the build container entrypoint
https://github.com/ae6rt/aftomato/blob/master/build-container/build.sh#L44

```
sh /home/aftomato/buildscripts/aftomato-build-scripts/${PROJECT_KEY}/build.sh 2>&1 | tee $CONSOLE
```

The build container will call your project's build script, capture
the console logs, and ship the build artifacts, console logs and
build metadata to S3 and DynamoDb.

## Build environment

The following environment variables are available in your build scripts:

* BUILD_ID:  UUID that uniquely identifies this build
* PROJECT_KEY: a composite key consisting of your project + repository
* BRANCH_TO_BUILD: an optional git branch to build within your application project. This is typically used with Github or Stash post commit hook events.

Concurrent builds of a given project + branch are currently forbidden,
and enforced with a lock in etcd, which also runs in the Aftomoto
cluster.
