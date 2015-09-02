#!/bin/bash

set -ux

for i in console-logs build-artifacts; do 
	aws --profile petrovic s3api create-bucket --bucket fosse-$i --create-bucket-configuration LocationConstraint=us-west-1
done

