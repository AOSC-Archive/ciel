package main

import "fmt"

func docHelp() {
	fmt.Print(`Usage:
	ciel init
	ciel load-os [TAR_FILE]    // unpack OS tarball or fetch the latest from internet directly
	ciel load-tree [GIT_URL]   // clone package tree from your link or AOSC OS ABBS at GitHub

	ciel update-os      // similar to 'apt update && apt full-upgrade'
	ciel update-tree    // similar to 'git pull'

	ciel [list]
	ciel add INSTANCE
	ciel del INSTANCE
	ciel shell -i INSTANCE         // start an interactive shell
	ciel shell -i INSTANCE "SHELL COMMAND LINE"
	ciel config -i INSTANCE        // configure toolchain for building (interactively)
	ciel build -i INSTANCE PACKAGE
	ciel rollback -i INSTANCE 

Rarely used:
	ciel mount [-i INSTANCE]   // mount all or one instance
	ciel down [-i INSTANCE]    // shutdown and unmount all or one instance
	ciel stop -i INSTANCE      // shutdown an instance
	ciel commit -i INSTANCE    // commit the change onto the underlying OS image
	ciel shell-with-poweroff -i INSTANCE "SHELL COMMAND LINE"
	                  // similar to 'shell' and additionally poweroff after executing command line,
	                  // may be useful for updating critical system components (dbus, systemd, etc.)
	ciel run -i INSTANCE ABSPATH_TO_EXE ARG1 ARG2 ...
	                  // lower-level version of 'shell', without login environment,
	                  // without sourcing ~/.bash_profile

Global flags:
	-C CIEL_DIR    // use CIEL_DIR as workdir instead of current directory
	-i INSTANCE    // specify the INSTANCE to manipulate
`)
}
