#!/bin/bash

git submodule update --init --recursive
make ffi
make

if [ ! -f "~/.swan/client/config.toml" ]; then
    cp ./config/config.toml.example ~/.swan/client/config.toml
    echo "~/.swan/client/config.toml created"
else
    echo "~/.swan/client/config.toml exists"
fi
