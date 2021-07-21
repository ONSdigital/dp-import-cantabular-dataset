#!/bin/bash -eux

pushd dp-import-cantabular-dataset
  COMPONENT_TEST_USE_LOG_FILE=true make test-component
  e=$?
  f="log-output.txt"
  cat $f && rm $f
popd
exit $e
