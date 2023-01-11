#!/bin/bash

CONF_FILE_DIR=${HOME}/.swan/client
mkdir -p ${CONF_FILE_DIR}

current_create_time=`date +"%Y%m%d%H%M%S"`

if [ -f "${CONF_FILE_DIR}/config.toml"  ]; then
    mv ${CONF_FILE_DIR}/config.toml  ${CONF_FILE_DIR}/config.toml.${current_create_time}
    echo "The previous configuration files have been backed up: ${CONF_FILE_DIR}/config.toml.${current_create_time}"
    cp ./config/config.toml.example ${CONF_FILE_DIR}/config.toml
    echo "${CONF_FILE_DIR}/config.toml created"
else
    cp ./config/config.toml.example ${CONF_FILE_DIR}/config.toml
    echo "${CONF_FILE_DIR}/config.toml created"
fi

if [ -f "${CONF_FILE_DIR}/chain-rpc.json"  ]; then
    mv ${CONF_FILE_DIR}/chain-rpc.json  ${CONF_FILE_DIR}/chain-rpc.json.${current_create_time}
    echo "The previous configuration files have been backed up: ${CONF_FILE_DIR}/chain-rpc.json.${current_create_time}"
    cp ./config/chain-rpc.json ${CONF_FILE_DIR}/chain-rpc.json
    echo "${CONF_FILE_DIR}/chain-rpc.json created"
else
    cp ./config/chain-rpc.json ${CONF_FILE_DIR}/chain-rpc.json
    echo "${CONF_FILE_DIR}/chain-rpc.json created"
fi

git submodule update --init --recursive
make ffi
make
make install-client