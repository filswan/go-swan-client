#!/bin/bash

CODE_URL=https://github.com/filswan/go-swan-client
BINARY_NAME=swan-client
TAG_NAME=v0.1.0-rc1

wget ${CODE_URL}/releases/download/${TAG_NAME}/${BINARY_NAME} --no-check-certificate
wget ${CODE_URL}/releases/download/${TAG_NAME}/config.toml.example --no-check-certificate

chmod +x ./${BINARY_NAME}

CONF_FILE_DIR=${HOME}/.swan/client
mkdir -p ${CONF_FILE_DIR}

CONF_FILE_PATH=${CONF_FILE_DIR}/config.toml
echo "config file path is: $CONF_FILE_PATH"

if [ -f "${CONF_FILE_PATH}" ]; then
    echo "${CONF_FILE_PATH} exists"
else
    cp ./config.toml.example $CONF_FILE_PATH
    echo "${CONF_FILE_PATH} created"
fi
