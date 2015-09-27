#!/bin/bash

set -ux

env | sort

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

   bctool s3put --aws-region ${AWS_DEFAULT_REGION} --bucket-name decap-build-artifacts --build-id ${BUILD_ID} --content-type application/x-gzip --filename /tmp/${TAR}.gz 
   bctool s3put --aws-region ${AWS_DEFAULT_REGION} --bucket-name decap-console-logs  --build-id ${BUILD_ID} --content-type application/x-gzip --filename ${CONSOLE}.gz

   bctool record-build-metadata  --aws-region ${AWS_DEFAULT_REGION}  --table-name decap-build-metadata  --build-id ${BUILD_ID}  --project-key ${PROJECT_KEY} --branch ${BRANCH_TO_BUILD} \
	--build-start-time ${START} --build-duration ${DURATION} --build-result ${BUILD_EXIT_CODE} 
else
   exec "$@"
fi
