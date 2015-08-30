#!/bin/bash

set -uvx

if [ $# -eq 0 ]; then

	cat <<EOF > $HOME/.aws/credentials
[default]
aws_access_key_id = $(cat /etc/secrets/aws-key)
aws_secret_access_key = $(cat /etc/secrets/aws-secret)
EOF

	TAR=archive.tar
	WORKSPACE=/home/aftomato/workspace
	CONSOLE=/tmp/console.log

let START=$(date +%s)

	pushd $WORKSPACE
	/home/aftomato/buildscripts/build-scripts/${PROJECT_KEY}/build.sh 2>&1 | tee $CONSOLE
	BUILD_EXITCODE=${PIPESTATUS[0]}
	popd

	let STOP=$(date +%s)
	DURATION=`expr $STOP - $START`

	# we need a way to configure the argument to -name.  'target' works for Javan/maven but is not sufficiently general for other types of builds.
	# In fact, for some builds (e.g., for interpreted language apps), there may be no build artifacts at all.
	TARGET=$(find $WORKSPACE -maxdepth 2 -name target -type d)
	tar cf $TAR $TARGET
	gzip $TAR 

	aws s3 cp ${TAR}.gz s3://beta-build-server-archive/$BUILD_ID
	aws s3 cp $CONSOLE s3://beta-build-server-console/$BUILD_ID

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
	LOCK_KEY=$(echo -n "${PROJECT_KEY}/${BRANCH_TO_BUILD}" | php -r "echo rawurlencode(fgets(STDIN));" 
	curl --cacert $CERT  -H"Authorization: Bearer $TOKEN" -i https://kubernetes/api/v1/proxy/namespaces/default/services/lockservice/v2/keys/${LOCK_KEY}?prevValue=${BUILD_ID} -XDELETE
else
	exec "$@"
fi
