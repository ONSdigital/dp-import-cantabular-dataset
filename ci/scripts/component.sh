#!/bin/bash -eux

pushd dp-import-cantabular-dataset
  COMPOSE_HTTP_TIMEOUT=240 COMPONENT_TEST_USE_LOG_FILE=true make test-component
  e=$?
  f="log-output.txt"
  cat $f && rm $f
popd

# Show message to prevent any confusion by 'ERROR 0' output
echo "please ignore error codes 0, like so: ERROR[xxxx] 0, as error code 0 means that there was no error"

# exit with the same code returned by docker compose
exit $e
