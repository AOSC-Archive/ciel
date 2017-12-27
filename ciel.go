// +build: linux
package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"ciel/internal/dotciel"
)

var rawArgs []string

func main() {
	var subCmd string
	if len(os.Args) >= 2 {
		if strings.HasPrefix(os.Args[1], "-") {
			rawArgs = os.Args[1:]
		} else {
			subCmd = os.Args[1]
			rawArgs = os.Args[2:]
		}
	} else {
		rawArgs = nil
	}
	flag.CommandLine = flag.NewFlagSet(os.Args[0]+" "+subCmd, flag.ExitOnError)
	router(subCmd)
}

func parse() {
	flag.CommandLine.Parse(rawArgs)
}

func router(subCmd string) {
	var routeTable = map[string]func(){
		// Create Directory Structures
		"init": initCiel, // here

		// Cleaning Up Before Shutting Down Computer
		"shutdown": shutdownCiel, // shutdown.go

		// Preparing and Removing Instance
		"add":      add,          // instances.go
		"load-os":  untarGuestOS, // guest_os.go
		"update":   update,       // guest_os.go
		"rollback": rollback,     // guest_os.go
		"del":      del,          // instances.go

		// Maintaining Instance Status
		"unlock":  unlockInst,  // shutdown.go
		"stop":    stop,        // instances.go
		"list":    list,        // status.go
		"mount":   mountCiel,   // mount_points.go
		"unmount": unmountCiel, // mount_points.go
		"":        list,        // status.go

		// Executing Commands
		"shell": shell, // instances.go
		"run":   run,   // instances.go

		//// Preparing Build
		//"clone":  clone,       // tree.go
		//"tree":   tree,        // tree.go
		//"config": buildConfig, // build.go
		//
		//// Building
		//"build": build, // build.go
	}
	requireEUID0()
	route, exists := routeTable[subCmd]
	if exists {
		route()
		return
	}
	plugins := getPlugins()
	for _, pluginItem := range plugins {
		if subCmd == pluginItem.Name {
			plugin(subCmd)
			return
		}
	}
	log.Fatalln("unknown command: " + subCmd)
}

func requireEUID0() {
	if os.Geteuid() != 0 {
		log.Fatalln("need to be root")
	}
}

func initCiel() {
	basePath := flagCielDir()
	parse()

	i := &dotciel.Ciel{BasePath: *basePath}
	i.Init()
}
