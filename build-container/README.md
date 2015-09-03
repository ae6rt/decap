## Base Build Container

This is the base build container for Decap.

We believe users will build their own containers using this as a
base, either retaining the entrypoint of the base or writing a new
one that extends it.

The ENTRYPOINT script build.sh executes a userland build script and
posts the results to S3 and DynamoDb.

## Building

To build and push the base build container, you will need to change
the container image name in the Makefile to point to your Docker
Hub username.  

To build and push:

```
make push
```
