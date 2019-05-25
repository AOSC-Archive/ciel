# CIEL 2 [![Build Status](https://api.travis-ci.org/AOSC-Dev/ciel.svg)](https://travis-ci.org/AOSC-Dev/ciel)
An **integrated packaging environment** for AOSC OS.

**Ciel** uses *systemd-nspawn* container as its backend and *overlay* file system as support rollback feature.

## Manual

```bash
ciel help
```

## Installation

```bash
make
sudo make install
```

You may use `make PREFIX=/usr` and `sudo make install PREFIX=/usr` to install to other location. Defaults to `/usr/local`.

## Dependencies

Building:
- Go
- C compiler
- make
- curl

Runtime:
- Systemd
- tar
- dos2unix

Runtime Kernel:
- Overlay file system
- System-V semaphores
