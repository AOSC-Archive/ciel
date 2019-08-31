// +build: linux
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"ciel/display"
	"ciel/internal/ciel"
)

var rawArgs []string

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
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
	switch subCmd {
	case "version":
		fmt.Println("Ciel", Version)
		return
	case "help":
		docHelp()
		return
	}
	var routeTable = map[string]func(){
		// Create Directory Structures
		"init":     initCiel, // here
		"farewell": farewell, // here

		// Preparing and Removing Instance
		"add":           add,          // instances.go
		"load-os":       untarGuestOS, // guest_os.go
		"factory-reset": factoryReset, // guest_os.go
		"update-os":     update,       // guest_os.go
		"rollback":      rollback,     // instances.go
		"commit":        commit,       // instances.go
		"del":           del,          // instances.go

		// Maintaining Instance Status
		"mount": mountCiel, // mount_points.go
		"stop":  stop,      // instances.go
		"down":  shutdown,  // mount_points.go
		"list":  list,      // status.go
		"":      list,      // status.go

		// Executing Commands
		"shell": shell, // instances.go
		"run":   run,   // instances.go

		// Building Configuration
		"load-tree":   clone,       // tree.go
		"update-tree": pull,        // tree.go
		"config":      buildConfig, // build.go

		// Building
		"build": build, // build.go

		"doctor": doctor, // doctor.go
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

	i := &ciel.Ciel{BasePath: *basePath}
	i.Init()
}

func farewell() {
	basePath := flagCielDir()
	batchFlag := flagBatch()
	parse()

	i := &ciel.Ciel{BasePath: *basePath}
	i.Check()
	c := i.Container()

	if !*batchFlag && d.ASKLower("DELETE ALL CIEL THINGS?", "yes/no") != "yes" {
		os.Exit(1)
	}

	d.SECTION("Farewell To Thee (Good Bye)")
	instList := c.GetAll()
	for _, inst := range instList {
		d.SECTION("Shutdown Instance " + inst.Name)
		inst.Unmount()
	}
	d.SECTION("DELETE .ciel DIRECTORY")
	d.ITEM("delete")
	err := os.RemoveAll(i.CielDir())
	d.ERR(err)
}
