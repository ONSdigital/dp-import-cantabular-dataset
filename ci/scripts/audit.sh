#!/bin/bash -eux

export cwd=$(pwd)

pushd $cwd/dp-import-cantabular-dataset
  make audit
popd