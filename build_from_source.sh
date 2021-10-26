#!/bin/bash

CONF_FILE_PATH=${HOME}/.swan/client/config.toml
echo $CONF_FILE_PATH

if [ -f "${CONF_FILE_PATH}" ]; then
    echo "~/.swan/client/config.toml exists"
else
    cp ./config/config.toml.example $CONF_FILE_PATH
    echo "~/.swan/client/config.toml created"
fi

git submodule update --init --recursive
make ffi
make

