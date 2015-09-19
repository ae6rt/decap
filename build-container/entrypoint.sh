#!/bin/bash

set -ux

env | sort

if [ $# -eq 0 ]; then

   TAR=archive.tar
   ARTIFACTS=/build-artifacts
   WORKSPACE=/home/decap/workspace
   CONSOLE=/tmp/console.log

   let START=$(date +%s)

   cat <<EOF > buildstart.json
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
        "branch": {
            "S": "$BRANCH_TO_BUILD"
        },
        "isBuilding": {
            "N": "1"
        }
}
EOF

   aws dynamodb put-item --table-name decap-build-metadata --item file://buildstart.json

   pushd $WORKSPACE

   sh /home/decap/buildscripts/decap-build-scripts/${PROJECT_KEY}/build.sh 2>&1 | tee $CONSOLE
   BUILD_EXITCODE=${PIPESTATUS[0]}

   popd

   pushd /build-artifacts
   tar czf /tmp/${TAR}.gz .
   popd

   gzip $CONSOLE

   let STOP=$(date +%s)
   DURATION=`expr $STOP - $START`

   aws s3 cp --content-type application/x-gzip /tmp/${TAR}.gz s3://decap-build-artifacts/$BUILD_ID
   aws s3 cp --content-type application/x-gzip ${CONSOLE}.gz s3://decap-console-logs/$BUILD_ID

   cat <<EOF > buildstop.json
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
    "buildResult": {
        "N": "$BUILD_EXITCODE"
    },
    "branch": {
        "S": "$BRANCH_TO_BUILD"
    },
    "isBuilding": {
        "N": "0"
    }
}
EOF

   aws dynamodb put-item --table-name decap-build-metadata --item file://buildstop.json
	
   curl -i http://lockservice.decap-system:2379/v2/keys/buildlocks/${BUILD_LOCK_KEY}?prevValue=${BUILD_ID} -XDELETE
else
   exec "$@"
fi
