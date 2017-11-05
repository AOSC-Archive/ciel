// +build: linux
package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"ciel/internal/cieldir.1"
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
	i := &cieldir.CielDir{}
	i.Check()
	var routeTable = map[string]func(){
		"init":              initCiel,
		"mount":             mountCiel,
		"umount":            unmountCiel,
		"unmount":           unmountCiel,
		"load":              unTar,
		"add":               addInst,
		"del":               delInst,
		"unlock-filesystem": unlockFileSystem,
		"unlock-container":  unlockContainer,
		"run":               run,
		"stop":              stop,
		"list":              list,
		"":                  list,
	}
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

func saveEnv(key, value string) {
	os.Setenv(key, value)
}
