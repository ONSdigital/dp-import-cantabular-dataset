#!/bin/bash -eux

echo "++++ component.sh starting..."
echo docker version
echo docker-compose version

pushd dp-import-cantabular-dataset
  make test-component
  # COMPONENT_TEST_USE_LOG_FILE=true make test-component
  e=$?
  # f="log-output.txt"
  # cat $f && rm $f
popd
exit $e

# Show message to prevent any confusion by 'ERROR 0' outpout
echo "please ignore error codes 0, like so: ERRO[xxxx] 0, as error code 0 means that there was no error"