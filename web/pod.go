package main

var podTemplate = `
{
    "kind": "Pod",
    "apiVersion": "v1",
    "metadata": {
        "name": "{{.BuildID}}",
        "namespace": "default",
        "creationTimestamp": null,
        "labels": {
            "type": "aftomato-build"
        }
    },
    "spec": {
        "volumes": [
            {
                "name": "build-scripts",
                "gitRepo": {
                    "repository": "{{.BuildScriptsGitRepo}}",
                    "revision": ""
                }
            },
            {
                "name": "aws-credentials",
                "secret": {
                    "secretName": "build-server-aws-credentials"
                }
            }
        ],
        "containers": [
            {
                "name": "build-server",
                "image": "{{.BuildImage}}",
                "env": [
                    {
                        "name": "BUILD_ID",
                        "value": "{{.BuildID}}"
                    },
                    {
                        "name": "PROJECT_KEY",
                        "value": "{{.ProjectKey}}"
                    },
                    {
                        "name": "BRANCH_TO_BUILD",
                        "value": "{{.BranchToBuild}}"
                    },
                    {
                        "name": "ARTIFACTS_BUCKET",
                        "value": "{{.BuildArtifactBucketName}}"
                    },
                    {
                        "name": "CONSOLE_LOGS_BUCKET",
                        "value": "{{.ConsoleLogsBucketName}}"
                    }
                ],
                "resources": {},
                "volumeMounts": [
                    {
                        "name": "build-scripts",
                        "mountPath": "/home/aftomato/buildscripts"
                    },
                    {
                        "name": "aws-credentials",
                        "mountPath": "/etc/secrets"
                    }
                ]
            }
        ],
        "restartPolicy": "Never"
    }
}
`
