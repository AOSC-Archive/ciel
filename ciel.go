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
		"init":              initCiel,         // init_mount_unmount.go
		"mount":             mountCiel,        // init_mount_unmount.go
		"umount":            unmountCiel,      // init_mount_unmount.go
		"unmount":           unmountCiel,      // init_mount_unmount.go
		"load":              unTar,            // load.go
		"add":               addInst,          // add_del.go
		"del":               delInst,          // add_del.go
		"unlock-filesystem": unlockFileSystem, // unlock.go
		"unlock-container":  unlockContainer,  // unlock.go
		"run":               run,              // run_stop.go
		"stop":              stop,             // run_stop.go
		"list":              list,             // status.go
		"":                  list,             // status.go
	}
	requireEUID0()
	route, exists := routeTable[subCmd]
	if exists {
		route()
		return
	}
	log.Fatalln("unknown command")
}

func getEnv(key, def string) string {
	v, ok := os.LookupEnv(key)
	if !ok {
		return def
	}
	return v
}

func requireEUID0() {
	if os.Geteuid() != 0 {
		log.Fatalln("need to be root")
	}
}