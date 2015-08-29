This is the base build container for aftomato.

We believe users will build their own containers using this as a
base.  Of course, they cannot tamper with the base image or entrypoint
script without breaking something.

The ENTRYPOING script build.sh executes a userland build script and
posts the results to S3 and DynamoDb.
