# ciel
A CI system for packaging.

## usage
```
commands:
	shell       shell in container
	run <cmd>   run command in container
	build <pkg> build package in container
	(default)   show this information
```

## compiling
Run `build.sh`, and `ciel` will be there.

## dependencies
- systemd's container components
  - `systemd-nspawn`
  - `machinectl`
- overlayfs (kernel module)
