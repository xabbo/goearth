#!/bin/bash

# Checks if the generated extensions build successfully

go install ../cmd/goearth
for client in {flash,shockwave}; do
    goearth new -c $client -d $client
    pushd $client
    # Add goearth & current module to workspace
    go work init ../.. .
    go build -v .
    popd
done