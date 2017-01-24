#!/bin/bash
set -euo pipefail
IFS=$'\n\t'
NAME=$1
HASH=$(git rev-parse --short HEAD)

# Push same image twice, once with the commit hash as the tag, and once with
# 'latest' as the tag. 'latest' will always refer to the last image that was
# built, since the next time this script is run, it'll get overridden. The
# commit hash, however, is a constant reference to this image.
docker tag -f $NAME $NAME:$HASH
docker push $NAME:$HASH
docker tag -f $NAME $NAME:latest
docker push $NAME:latest
