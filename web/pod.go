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
                    "secretName": "aftomato-aws-credentials"
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
                        "name": "BUILD_LOCK_KEY",
                        "value": "{{.BuildLockKey}}"
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
