# ciel
A CI system for packaging.

## usage
```
ciel-cli help
```

## compiling
Install `Go`, run `build.sh`, and `ciel-cli` will be there.

## dependencies
- systemd's container components
  - `systemd-nspawn`
  - `machinectl`
- overlayfs (kernel module)
- coreutils
- tar
