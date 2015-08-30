## Overview

A CI build server based on a Kubernetes backend that executes builds in a specially prepared container.

## AWS Setup

### Install the AWS CLI app

See http://aws.amazon.com/documentation/cli/

Aftomato uses S3 buckets to store build artifacts and console logs, and DynamoDb to store overall build results.

### IAM user

Aftomato stores build information in S3 buckets and a DynamoDb table.  These buckets and table are secured using
AWS access policies that are associated with a dedicated IAM user named _aftomato_.

Create the IAM user named _aftomato_ in your AWS Console.  This
user needs no SSH private key, but it will need the usual AWS Access
Key and associated Secret.  When you create the user, AWS Console will display the access key and secret.  Save these
to a file --- the Aftomato build container will need them to publish build results, and the web frontend to Aftomato will
need them to perform build queries and administration.

View the new user in your AWS Console and note its Amazon Resource Name (ARN).  It will look like this

```
User ARN: arn:aws:iam::<your account ID>:user/aftomato

```

where _your account ID_ is your Amazon account ID.

### Buckets

Create two S3 buckets, one for build artifacts (even if you don't think you'll have any) and build console logs.

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

Select the bucket of interest in the AWS S3 Console area, and attach these policies to their respective buckets.

### DynamoDb

In your AWS Console create a DynamoDb table named _aftomato-build-metadata_
in your preferred region that has these properties.  The table
should have a main hashkey name _buildID_ with no range key, and a
global secondary index with hashkey _projectKey_ and range _buildTime_.

N.B. You are responsible for getting the _ProvisionedThroughput_
right based on your anticipated usage, which is related to how much
Amazon will bill you for that usage.

This an example, but it shows the highlights:

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

In the IAM Policy section of your AWS Console create a policy named _aftomato-db_ that looks like

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
        },
        {
            "Sid": "<other statement ID>",
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

adjusting the AWS region as appropriate to where you created the DynamoDb table.

In the Users section of the AWS Console, attach this _aftomato-db_ policy to the _aftomato_ user.

