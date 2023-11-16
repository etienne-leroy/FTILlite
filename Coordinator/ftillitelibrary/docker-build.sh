#!/bin/bash
set -e

GIT_COMMIT="$(git describe --always --tags --long --dirty)"

docker build . \
    -t "fintracerpoc-frontend:latest" \
    -t "fintracerpoc-frontend:${GIT_COMMIT}"