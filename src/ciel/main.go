package main

import (
	"ciel-driver"
	"log"
	"math/rand"
	"os"
	"os/exec"
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
		cmd := exec.Command("ciel-"+command, args...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		// FIXME: exit code
		if err != nil {
			os.Exit(1)
		}
	case "help":
		printHelp()
	case "shell":
		c := ciel.New(CielMachineName, CielRoot)
		exitcode := c.Shell()
		c.Shutdown()
		c.Fs.Unmount()
		os.Exit(exitcode)
	case "init":
		panicIfNotEnough(args, 1)
		err := genesis(args[0], CielRoot)
		if err != nil {
			log.Fatalln(err)
		}
	case "stub-upd":
		c := ciel.New(CielMachineName, CielRoot)
		err := updateStub(c)
		c.Shutdown()
		c.Fs.Unmount()
		if err != nil {
			log.Fatalln(err)
		}
	case "dist-cfg":
		c := ciel.New(CielMachineName, CielRoot)
		err := configDist(c)
		c.Shutdown()
		c.Fs.Unmount()
		if err != nil {
			log.Fatalln(err)
		}
	case "dist-upd":
		c := ciel.New(CielMachineName, CielRoot)
		err := updateDist(c)
		c.Shutdown()
		c.Fs.Unmount()
		if err != nil {
			log.Fatalln(err)
		}
	case "clean":
		c := ciel.New(CielMachineName, CielRoot)
		err := cleanRelease(c)
		c.Shutdown()
		c.Fs.Unmount()
		if err != nil {
			log.Fatalln(err)
		}
	}
}

func panicIfNotEnough(a []string, c int) {
	if len(a) != c {
		panic("argument count mismatch")
	}
}

func init() {
	rand.Seed(time.Now().UnixNano())
	ciel.FileSystemLayers = ciel.Layers{
		"99-upperdir",
		"80-cache",
		"50-override",
		"10-dist",
		"00-stub",
	}
}
