#!/bin/sh
# Entrypoint script for running Go lambda function in container.

if [ -z "${AWS_LAMBDA_RUNTIME_API}" ]; then
  exec /usr/local/bin/aws-lambda-rie $@
else
  exec $@
fi     
