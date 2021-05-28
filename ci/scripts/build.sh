#!/bin/bash -eux

pushd dp-import-cantabular-dataset
  make build
  cp build/dp-import-cantabular-dataset Dockerfile.concourse ../build
popd
