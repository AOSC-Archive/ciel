package main

import (
	"ci"
	"fmt"
	"log"
	"os"
)

const MachineName = "buildkit"
const ShellPath = "/bin/bash"
const ACBSPath = "/bin/acbs-build"

type HandlerFunc func(*ci.ContainerInstance, []string) error
type Route struct {
	name    string
	argc    int // -1 = infinite
	needfs  bool
	needci  bool
	handler HandlerFunc
}

var router = []Route{
	{"init", 1, true, true, cielpostinit},
	{"update", 0, true, true, cielupdate},
	{"shell", 0, true, true, cielshell},
	{"run", -1, true, true, cielrun},
	{"build", 1, true, true, cielbuild},
	{"tbuild", 1, true, true, cieltbuild},
	{"rbuild", 1, true, true, cielrbuild},
	{"collect", 0, false, false, cielcollect},
	{"clean", 0, false, false, cielclean},
}

func main() {
	var route *Route

	if len(os.Args) == 1 {
		cielhelp()
		return
	}

	for _, xroute := range router {
		if os.Args[1] == xroute.name && (xroute.argc == -1 || len(os.Args)-2 == xroute.argc) {
			route = &xroute
			break
		}
	}
	if route == nil {
		cielhelp()
		return
	}

	var cielInitStage = 0

	if os.Args[1] == "init" && len(os.Args)-2 == 1 {
		cielInitStage = 1
	}

BEGIN_MAIN:

	fs := ci.InitFilesystem("container-layers") // FIXME: configurability
	container := ci.NewContainer(fs, MachineName)

	switch cielInitStage {
	case 1:
		if err := cielinitAptUpdate(container, os.Args[2]); err != nil {
			log.Panicln(err)
		}
	case 2:
		if err := cielinitAptInstallSystemd(container, os.Args[2]); err != nil {
			log.Panicln(err)
		}
	}

	if route.needfs {
		if err := fs.Startup(); err != nil {
			log.Panicln("fs", err)
		}
		defer func() {
			if err := fs.Shutdown(); err != nil {
				log.Panicln("shutdownfs", err)
			}
		}()

		if route.needci {
			if err := container.Startup(); err != nil {
				log.Panicln("container", err)
			}
			defer func() {
				if err := container.Shutdown(); err != nil {
					log.Panicln("shutdown", err)
				}
			}()
		}
	}

	switch cielInitStage {
	case 1:
		cielInitStage = 2
		goto BEGIN_MAIN
	case 2:
		cielInitStage = 3
		goto BEGIN_MAIN
	}

	args := []string{}
	if len(os.Args) > 2 {
		args = os.Args[2:]
	}
	if err := route.handler(container, args); err != nil {
		log.Panicln(err)
	}
}

func cielhelp() {
	fmt.Print(`Usage:	` + os.Args[0] + ` <command> [arg...]

Commands:

	Preparing:
		init    <tar>   Bootstrap from "stub" tarball
		update          Update packages for BuildKit layer

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

Filesystem:

	(Current directory)
	|
	|-- container-layers/
	| |-- 00-stub/           extracted from tarball
	| |-- 01-stub-config/    /etc/apt will be merged to here
	| |-- 10-buildkit/
	| |-- 50-cache/          /var/cache will be merged to here
	| |-- 99-upperdir/       the surface of filesystem
	| \-- 99-upperdir-work/  temporary directory for overlayfs
	|
	\-- collected/
	  |-- report.tar.xz
	  \-- pkg/
	    \-- os-amd64/
	      \-- os3-dpkg/
	        \-- ?/
	          \-- ?????.deb

`)
}
