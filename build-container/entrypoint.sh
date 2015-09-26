#!/bin/bash

set -ux

env | sort

if [ $# -eq 0 ]; then

   TAR=archive.tar
   ARTIFACTS=/build-artifacts
   WORKSPACE=/home/decap/workspace
   CONSOLE=/tmp/console.log

   START=$(date +%s)

   bctool build-start --aws-region ${AWS_DEFAULT_REGION} --table-name decap-build-metadata --build-id ${BUILD_ID} --start-time ${START} --project-key ${PROJECT_KEY} --branch ${BRANCH_TO_BUILD} 

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

   bctool build-finish --aws-region ${AWS_DEFAULT_REGION} --table-name decap-build-metadata --build-id ${BUILD_ID} --build-duration ${DURATION} --build-result ${BUILD_EXIT_CODE} 

   curl -i http://lockservice.decap-system:2379/v2/keys/buildlocks/${BUILD_LOCK_KEY}?prevValue=${BUILD_ID} -XDELETE
else
   exec "$@"
fi
