#!/bin/bash -eux

pushd dp-import-cantabular-dataset
  make test-component
popd
