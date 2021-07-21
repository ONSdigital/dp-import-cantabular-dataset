#!/bin/bash -eux

pushd dp-import-cantabular-dataset
  make test-component
  e=$?
  cat log-output.txt && rm log-output.txt
popd
exit $e
