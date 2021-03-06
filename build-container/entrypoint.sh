#!/bin/bash

set -ux

env | sort

readonly AWS_ACCESS_KEY_ID=$(cat /etc/secrets/aws-key)
readonly AWS_SECRET_ACCESS_KEY=$(cat /etc/secrets/aws-secret)
readonly AWS_DEFAULT_REGION=$(cat /etc/secrets/aws-region)

if [ $# -eq 0 ]; then

   TAR=archive.tar
   ARTIFACTS=/build-artifacts
   WORKSPACE=/home/decap/workspace
   CONSOLE=/tmp/console.log

   START=$(date +%s)

   pushd $WORKSPACE
   	sh /home/decap/buildscripts/decap-build-scripts/${PROJECT_KEY}/build.sh 2>&1 | tee $CONSOLE
   	BUILD_EXITCODE=${PIPESTATUS[0]}
   popd

   STOP=$(date +%s)
   DURATION=$((STOP - START))

   gzip $CONSOLE

   pushd ${ARTIFACTS}
   	tar czf /tmp/${TAR}.gz .
   popd

   # Store the build artifacts
   bctool s3put --aws-access-key-id ${AWS_ACCESS_KEY_ID} --aws-secret-access-key ${AWS_SECRET_ACCESS_KEY} --aws-region ${AWS_DEFAULT_REGION} \
	--bucket-name decap-build-artifacts --build-id ${BUILD_ID} --content-type application/x-gzip --filename /tmp/${TAR}.gz 

   # Store the console log
   bctool s3put --aws-access-key-id ${AWS_ACCESS_KEY_ID} --aws-secret-access-key ${AWS_SECRET_ACCESS_KEY} --aws-region ${AWS_DEFAULT_REGION} \
	--bucket-name decap-console-logs  --build-id ${BUILD_ID} --content-type application/x-gzip --filename ${CONSOLE}.gz

   # Store the build metadata
   bctool record-build-metadata  --aws-access-key-id ${AWS_ACCESS_KEY_ID} --aws-secret-access-key ${AWS_SECRET_ACCESS_KEY} --aws-region ${AWS_DEFAULT_REGION} \
	--table-name decap-build-metadata  --build-id ${BUILD_ID}  --project-key ${PROJECT_KEY} --branch ${BRANCH_TO_BUILD} \
	--build-start-time ${START} --build-duration ${DURATION} --build-result ${BUILD_EXITCODE} 

else
   exec "$@"
fi
