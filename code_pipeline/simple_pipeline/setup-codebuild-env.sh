#!/bin/bash

# Install go 1.17.7
INSTALL_PATH="/usr/local"
INSTALL_VERSION="1.17.7"

export GOROOT="${INSTALL_PATH}/go"
export PATH=$GOROOT/bin:$PATH

ARCH="$(uname -m)"
FILE_NAME="go${INSTALL_VERSION}.linux-amd64.tar.gz"
# Is arm64 machine?
if [ "$ARCH" = "aarch64" ]; then
    FILE_NAME="go${INSTALL_VERSION}.linux-arm64.tar.gz"
fi
FILE_URL="https://dl.google.com/go/${FILE_NAME}"

pushd /tmp
yum -y install tar gzip
curl $FILE_URL -o $FILE_NAME
rm -rf $GOROOT && tar -C $INSTALL_PATH -xzf ./$FILE_NAME
rm -rf ./$FILE_NAME
popd
