#!/bin/sh

cd filesystem/

if [ -d "overlayfs" ]; then
    umount overlayfs/mergeddir
    rm -rf overlayfs 
fi

mkdir -p overlayfs/lowerdir overlayfs/upperdir overlayfs/workdir overlayfs/mergeddir

for i in $(seq 1 3); do
    mkdir -p overlayfs/lowerdir/image-layer$i
    echo "This is layer $i" > overlayfs/lowerdir/image-layer$i/layer$i.txt
done

echo "This is the upper layer" > overlayfs/upperdir/layer.txt

mount -t overlay overlay -o lowerdir=overlayfs/lowerdir/image-layer1:overlayfs/lowerdir/image-layer2:overlayfs/lowerdir/image-layer3,upperdir=overlayfs/upperdir,workdir=overlayfs/workdir overlayfs/mergeddir