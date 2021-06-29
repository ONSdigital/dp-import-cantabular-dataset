#!/bin/bash -eux

cwd=$(pwd)

pushd $cwd/dp-import-cantabular-dataset
  make lint
popd
