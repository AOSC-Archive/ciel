package main

import (
	"ciel-driver"
	"log"
	"math/rand"
	"os"
	"time"
)

const (
	CielMachineName = "ciel"
	CielRoot        = "./cielfs"
)

var (
	DefaultCommand     = "shell"
	DefaultCommandArgs = []string{}
)

func main() {
	args := os.Args
	if len(os.Args) == 1 {
		router(DefaultCommand, DefaultCommandArgs)
	} else {
		router(args[1], args[2:])
	}
}

func router(command string, args []string) {
	switch command {
	default:
		printHelp()
		os.Exit(127)
	case "help":
		printHelp()
	case "shell":
		c := ciel.New(CielMachineName, CielRoot)
		exitcode := c.Shell()
		c.Fs.Unmount()
		c.Shutdown()
		os.Exit(exitcode)
	case "init":
		panicIfNotEnough(args, 1)
		err := genesis(args[0], CielRoot)
		if err != nil {
			log.Fatalln(err)
		}
	case "stub-update":
		c := ciel.New(CielMachineName, CielRoot)
		err := updateStub(c)
		if err != nil {
			log.Fatalln(err)
		}
	}
}

func panicIfNotEnough(a []string, c int) {
	if len(a) != c {
		panic("argument count mismatched")
	}
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
