## AWS Resource creation

Decap uses S3 buckets to store build artifacts and console logs,
and DynamoDb to store overall build results and metadata.

### Install the AWS CLI app

See http://aws.amazon.com/documentation/cli/

### IAM user

Decap stores build information in S3 buckets and a DynamoDb
table.  These buckets and table are secured using AWS access policies
that are associated with a dedicated IAM user named _decap_.

In this directory we provide Decap scripts for creating all the AWS
resources Decap needs.  To run these scripts effectively, you will
need an AWS account with what we're calling _root like_ powers.
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
* two DynamoDb tables, one named decap-build-metadata and another named decap-buildlocks
* six policies attached to the user decap

The six policies are named:

* decap-locks-base
* decap-locks-expires-index
* decap-db-base
* decap-db-project-key
* decap-s3-build-artifacts
* decap-s3-console-logs

The create-world.sh script also creates an AWS access key for the
user decap.  The key and secret are written to the file aws.credentials.
Use this access key and secret to craft the Kubernetes decap-secrets.yaml
file described in the main README of this project.

