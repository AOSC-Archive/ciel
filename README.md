# ciel
A CI system for packaging.

## usage
```
Usage: ciel <command> [arg...]

Most used commands:
(default)       Show this information
shell           Shell in container
run     <cmd>   Run command in container
build   <pkg>   Equivalent to "rbuild + collect + clean"
tbuild  <pkg>   Build package, but stay in shell to test the package

Underlying operations:
rbuild  <pkg>   Build package
collect         Collect packaging output and log files
clean           Merge cache to "overlay" and reset container
```
## compiling
Install `Go`, run `build.sh`, and `ciel` will be there.

## dependencies
- systemd's container components
  - `systemd-nspawn`
  - `machinectl`
- overlayfs (kernel module)
