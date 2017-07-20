# ciel [![Build Status](https://api.travis-ci.org/AOSC-Dev/ciel.svg)](https://travis-ci.org/AOSC-Dev/ciel)
A tool for controlling multi-layer file systems and containers

## manual

During the rapid iteration (before version 1.x.x), you may read help messages:
```
ciel help
```

Or Wiki:
- https://github.com/AOSC-Dev/ciel/wiki/The-Ciel-User-Manual-en
- https://github.com/AOSC-Dev/ciel/wiki/The-Ciel-User-Manual-zh_CN

## installation


```
git clone https://github.com/AOSC-Dev/ciel
cd ciel
git submodule update --init --recursive
make
sudo make install
```

## dependencies

Building:
- git
- make
- Go

Runtime:
- systemd's container components
- overlayfs (kernel module)
- coreutils
- tar
