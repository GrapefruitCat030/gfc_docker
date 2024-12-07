#!/bin/sh

cd filesystem/

if [ -d "aufs" ]; then
    rm -rf aufs
fi

mkdir -p aufs/container-layer aufs/mnt
for i in 1 2 3; do
    mkdir -p aufs/image-layer$i
    echo "image-layer$i" > aufs/image-layer$i/layer$i.txt
done

mount -t aufs -o dirs=aufs/container-layer=rw:aufs/image-layer1=ro:aufs/image-layer2=ro:aufs/image-layer3=ro none aufs/mnt
