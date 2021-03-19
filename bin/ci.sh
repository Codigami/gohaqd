#!/usr/bin/env bash

# Build the code, run the tests and
# create the "production" docker image - not necessarily in that order
#
# IMPORTANT: The docker image must be named <repo_name>:latest

docker build -t gohaqd .
