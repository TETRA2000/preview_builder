#!/usr/bin/env bash

PR_NO=$1
CD=$PWD

cd repo && \
    git clean -f
    git fetch --prune && \
    git checkout origin/master && \
    git pull --no-edit origin pull/${PR_NO}/head
    git gc && \
    cd ${CD}

docker build -t angular_io:${PR_NO} -f Dockerfile .
