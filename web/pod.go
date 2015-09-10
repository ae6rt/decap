package main

var podTemplate = `
{
    "kind": "Pod",
    "apiVersion": "v1",
    "metadata": {
        "name": "{{.BuildID}}",
        "namespace": "decap",
        "creationTimestamp": null,
        "labels": {
            "type": "decap-build",
            "parent": "{{.Parent}}",
            "library": "{{.Library}}",
            "branch": "{{.BranchToBuild}}",
        }
    },
    "spec": {
        "volumes": [
            {
                "name": "build-scripts",
                "gitRepo": {
                    "repository": "{{.BuildScriptsGitRepo}}",
                    "revision": "{{.BuildScriptsGitRepoBranch}}"
                }
            },
            {
                "name": "decap-credentials",
                "secret": {
                    "secretName": "decap-credentials"
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
                    },
                    {
                        "name": "AWS_ACCESS_KEY_ID",
                        "value": "{{.AWSAccessKeyID}}"
                    },
                    {
                        "name": "AWS_SECRET_ACCESS_KEY",
                        "value": "{{.AWSAccessSecret}}"
                    },
                    {
                        "name": "AWS_DEFAULT_REGION",
                        "value": "{{.AWSRegion}}"
                    }
                ],
                "resources": {},
                "volumeMounts": [
                    {
                        "name": "build-scripts",
                        "mountPath": "/home/decap/buildscripts"
                    },
                    {
                        "name": "decap-credentials",
                        "mountPath": "/etc/secrets"
                    }
                ]
            }
            {{.SidecarContainers}}
        ],
        "restartPolicy": "Never"
    }
}
`
