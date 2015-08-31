#!/bin/bash

set -ux

if [ $# -eq 0 ]; then
    mkdir -p $HOME/.aws
	cat <<EOF > $HOME/.aws/credentials
[default]
aws_access_key_id = $(cat /etc/secrets/aws-key)
aws_secret_access_key = $(cat /etc/secrets/aws-secret)
region = $(cat /etc/secrets/region)
EOF

	TAR=archive.tar
	WORKSPACE=/home/aftomato/workspace
	CONSOLE=/tmp/console.log

let START=$(date +%s)

	pushd $WORKSPACE

	/home/aftomato/buildscripts/build-scripts/${PROJECT_KEY}/build.sh 2>&1 | tee $CONSOLE
	BUILD_EXITCODE=${PIPESTATUS[0]}

	# todo what gets archived needs to be configurable
	tar czf /tmp/${TAR}.gz .

	popd

	gzip $CONSOLE

	let STOP=$(date +%s)
	DURATION=`expr $STOP - $START`

	aws s3 cp /tmp/${TAR}.gz s3://aftomato-build-artifacts/$BUILD_ID
	aws s3 cp ${CONSOLE}.gz s3://aftomato-console-logs/$BUILD_ID

	if [ $BUILD_EXITCODE -eq 0 ]; then
	     BUILD_STATUS="true"
	else
	     BUILD_STATUS="false"
	fi

	cat <<XXX > dynamodb.json
{
    "buildID": {
        "S": "$BUILD_ID"
    },
    "buildTime": {
        "N": "$START"
    },
    "projectKey": {
        "S": "$PROJECT_KEY"
    },
    "buildElapsedTime": {
        "N": "$DURATION"
    },
    "buildStatus": {
        "BOOL": "$BUILD_STATUS"
    },
    "branch": {
        "S": "$BRANCH_TO_BUILD"
    }
}
XXX

	aws dynamodb put-item --table-name aftomato-build-metadata --item file://dynamodb.json
	
	CERT=/var/run/secrets/kubernetes.io/serviceaccount/ca.crt
	TOKEN=$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)
	LOCK_KEY=$(echo -n "${PROJECT_KEY}/${BRANCH_TO_BUILD}" | php -r "echo rawurlencode(fgets(STDIN));")
	curl --cacert $CERT  -H"Authorization: Bearer $TOKEN" -i https://kubernetes/api/v1/proxy/namespaces/default/services/lockservice/v2/keys/${LOCK_KEY}?prevValue=${BUILD_ID} -XDELETE
else
	exec "$@"
fi
