// +build: linux
package main

import (
	"flag"
	"log"
	"os"
	"strings"
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
		"init":    initCiel,    // init_mount_unmount.go
		"mount":   mountCiel,   // init_mount_unmount.go
		"umount":  unmountCiel, // init_mount_unmount.go
		"unmount": unmountCiel, // init_mount_unmount.go
		"load":    unTar,       // load.go
		"add":     addInst,     // add_del.go
		"del":     delInst,     // add_del.go
		"unlock":  unlockInst,  // unlock.go
		"shell":   shell,       // run_stop.go
		"run":     run,         // run_stop.go
		"stop":    stop,        // run_stop.go
		"list":    list,        // status.go
		"":        list,        // status.go
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