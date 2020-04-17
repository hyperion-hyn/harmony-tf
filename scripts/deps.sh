#!/usr/bin/env bash

echo "Updating dependencies"

deps=(
harmony-one/harmony
harmony-one/go-sdk
harmony-one/go-lib
)

for dep in "${deps[@]}"; do
  echo "Updating dependency ${dep} to latest master"
  go get -u github.com/${dep}@master
done
