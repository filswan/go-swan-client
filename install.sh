#!/bin/bash

URL_PREFIX=https://github.com/filswan/go-swan-client/releases/download
BINARY_NAME=swan-client-2.3.0-linux-amd64
TAG_NAME=v2.3.0
wget --no-check-certificate ${URL_PREFIX}/${TAG_NAME}/${BINARY_NAME}
wget --no-check-certificate ${URL_PREFIX}/${TAG_NAME}/config.toml.example
wget --no-check-certificate ${URL_PREFIX}/${TAG_NAME}/chain-rpc.json

sudo install -C ${BINARY_NAME} /usr/local/bin/${BINARY_NAME}

CONF_FILE_DIR=${HOME}/.swan/client
mkdir -p ${CONF_FILE_DIR}

current_create_time=`date +"%Y%m%d%H%M%S"`

if [ -f "${CONF_FILE_DIR}/config.toml"  ]; then
    # shellcheck disable=SC2154
    mv ${CONF_FILE_DIR}/config.toml  ${CONF_FILE_DIR}/config.toml.${current_create_time}
    echo "The previous configuration files have been backed up: ${CONF_FILE_DIR}/config.toml.${current_create_time}"
    cp ./config.toml.example ${CONF_FILE_DIR}/config.toml
    echo "${CONF_FILE_DIR}/config.toml created"
else
    cp ./config.toml.example ${CONF_FILE_DIR}/config.toml
    echo "${CONF_FILE_DIR}/config.toml created"
fi

if [ -f "${CONF_FILE_DIR}/chain-rpc.json"  ]; then
    mv ${CONF_FILE_DIR}/chain-rpc.json  ${CONF_FILE_DIR}/chain-rpc.json.${current_create_time}
    echo "The previous configuration files have been backed up: ${CONF_FILE_DIR}/chain-rpc.json.${current_create_time}"
    cp ./chain-rpc.json ${CONF_FILE_DIR}/chain-rpc.json
    echo "${CONF_FILE_DIR}/chain-rpc.json created"
else
    cp ./chain-rpc.json ${CONF_FILE_DIR}/chain-rpc.json
    echo "${CONF_FILE_DIR}/chain-rpc.json created"
fi

chmod +x ./${BINARY_NAME}
