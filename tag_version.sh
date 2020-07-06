#!/bin/sh

VERSION=$1

if [ -z "$VERSION" ]; then
    echo "version must be set."
    exit 1
fi

echo $VERSION > ./VERSION
