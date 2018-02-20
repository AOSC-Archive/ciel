# ciel [![Build Status](https://api.travis-ci.org/AOSC-Dev/ciel.svg)](https://travis-ci.org/AOSC-Dev/ciel)
An integrated packaging environment for AOSC OS.

## manual

During the rapid iteration (before version 1.x.x), you may read help messages:
```bash
ciel help
```

## installation


```bash
git clone https://github.com/AOSC-Dev/ciel
cd ciel

make
sudo make install
```

## dependencies

Building:
- Go
- make

Runtime:
- systemd
- overlayfs (kernel feature)
- coreutils
- tar
