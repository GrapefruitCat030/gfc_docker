#!/bin/sh

if [ ! -d filesystem/busyboxfs ]; then
    echo "Filesystem not found, extracting..."
    mkdir -p filesystem/busyboxfs
    tar -C filesystem/busyboxfs -xvf filesystem/busyboxfs.tar
fi
