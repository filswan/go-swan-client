#!/bin/bash

CONF_FILE_DIR=${HOME}/.swan/client
mkdir -p ${CONF_FILE_DIR}

CONF_FILE_PATH=${CONF_FILE_DIR}/config.toml
echo $CONF_FILE_PATH

if [ -f "${CONF_FILE_PATH}" ]; then
    echo "${CONF_FILE_PATH} exists"
else
    cp ./config.toml.example $CONF_FILE_PATH
    echo "${CONF_FILE_PATH} created"
fi

git submodule update --init --recursive
make ffi
make

