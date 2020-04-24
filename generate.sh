#!/bin/bash

SRCDIR="$(dirname "$0")"
cat << EOF > config.go
// THIS FILE IS AUTO-GENERATED. DO NOT EDIT!
package main

// Automatically generated version and installation location
const (
        Version = "$(git describe --tags)"
        Prefix  = "$(readlink -f "${SRCDIR}")"
)
EOF

