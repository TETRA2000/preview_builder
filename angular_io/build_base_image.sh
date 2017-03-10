#!/usr/bin/env bash

CD=$PWD

cd repo && \
    git clean -f
    git fetch --prune && \
    git checkout origin/master && \
    git gc && \
    cd ${CD}

docker build -t angular_io_base -f Dockerfile.base .
