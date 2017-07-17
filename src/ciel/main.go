package main

import (
	"ciel-driver"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

const (
	MachineName   = "ciel"
	FileSystemDir = "./cielfs"
	LibExecDir    = "/usr/libexec"
)

var (
	DefaultCommand     = "shell"
	DefaultCommandArgs = []string{}
)

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
	case "init":
		cielInit(args)

	case "drop":
		cielDrop(args)
	case "mount":
		cielMount(args)
	case "merge":
		cielMerge(args)
	case "clean":
		cielClean(args)

	case "shell":
		cielShell(args)
	case "rawcmd":
		cielRawcmd(args)

	case "help":
		cielHelp(args)

	default:
		proc := LibExecDir + "/ciel-plugin/ciel-" + command
		cmd := exec.Command(proc, args...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				exitStatus := exitError.Sys().(syscall.WaitStatus)
				os.Exit(exitStatus.ExitStatus())
			} else {
				log.Fatalf("plugin %s not found\n", proc)
			}
		}
	}
}

// ciel init <tarball>
func cielInit(args []string) {
	if len(args) != 1 {
		log.Fatalln("init: you may only input one argument")
	}
	err := genesis(args[0], FileSystemDir)
	if err != nil {
		log.Fatalln(err)
	}
}

// ciel drop [<layers>]
func cielDrop(args []string) {
	c := ciel.New(MachineName, FileSystemDir)
	if len(args) == 0 {
		args = []string{"upperdir"}
	}
	for _, layer := range args {
		path := c.Fs.Layer(layer)
		if path == "" {
			log.Printf("drop: layer %s not exist\n", layer)
			continue
		}
		if err := os.RemoveAll(c.Fs.Layer(layer)); err != nil {
			log.Println(err)
		}
	}
}

// ciel mount [--read-write] [<layers>]
func cielMount(args []string) {
	c := ciel.New(MachineName, FileSystemDir)
	var rw = false
	if len(args) >= 1 && args[0] == "--read-write" {
		args = args[1:]
		rw = true
	}
	if len(args) > 0 {
		c.Fs.DisableAll()
		c.Fs.EnableLayer(args...)
	}
	var err error
	if rw {
		err = c.Fs.Mount()
	} else {
		err = c.Fs.MountReadOnly()
	}
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(c.Fs.TargetDir())
}

// ciel merge [<upper>..]<lower> [--no-self] path
func cielMerge(args []string) {
	// FIXME: limit arguments
	layers := strings.SplitN(args[0], "..", 2)
	if len(layers) == 1 { // "xx" => ["upperdir" "xx"]
		layers = append([]string{"upperdir"}, layers[0])
	} else if layers[0] == "" { // "..xx" => ["upperdir" "xx"]
		layers[0] = "upperdir"
	}
	var excludeSelf = false
	var path string
	excludeSelf = args[1] == "--no-self"
	if args[1] == "--no-self" {
		path = args[2]
	} else {
		path = args[1]
	}
	c := ciel.New(MachineName, FileSystemDir)
	c.Fs.MergeFile(path, layers[0], layers[1], excludeSelf)
}

// ciel clean [--factory-reset]
func cielClean(args []string) {
	c := ciel.New(MachineName, FileSystemDir)
	c.Fs.DisableLayer("override", "cache")
	var err error
	if len(args) == 1 && args[0] == "--factory-reset" {
		err = cleanFactoryReset(c)
	} else {
		err = cleanNormal(c)
	}
	c.Shutdown()
	c.Fs.Unmount()
	if err != nil {
		log.Fatalln(err)
	}
}

// ciel shell [<cmdline>]
func cielShell(args []string) {
	c := ciel.New(MachineName, FileSystemDir)
	var exitcode int
	if len(args) == 0 {
		exitcode = c.Shell()
	} else if len(args) == 1 {
		exitcode = c.Command(args[0])
	} else {
		log.Fatalln("shell: you may only input one argument")
	}
	c.Shutdown()
	c.Fs.Unmount()
	os.Exit(exitcode)
}

// ciel rawcmd <cmd> <arg1> <arg2> ...
func cielRawcmd(args []string) {
	if len(args) == 0 {
		log.Fatalln("init: you must input one argument at least")
	}
	c := ciel.New(MachineName, FileSystemDir)
	exitcode := c.CommandRaw(args[0], os.Stdin, os.Stdout, os.Stderr, args[1:]...)
	c.Shutdown()
	c.Fs.Unmount()
	os.Exit(exitcode)
}

func cielHelp(args []string) {
	fmt.Println("Usage: " + os.Args[0] + " [command [...]]")
	fmt.Println(`Default command is "shell".`)
	fmt.Println("")

	fmt.Println("Commands:")
	fmt.Println("\thelp")
	fmt.Println("")
	fmt.Println("\tinit   <tarball>")
	fmt.Println("")
	fmt.Println("\tdrop   [<layers>]")
	fmt.Println("\tmount  [--read-write] [<layers>]")
	fmt.Println("\tmerge  [<upper>..]<lower> [--no-self] path")
	fmt.Println("\tclean  [--factory-reset]")
	fmt.Println("")
	fmt.Println("\tshell  <cmdline>")
	fmt.Println("\trawcmd <cmd> <arg1> <arg2> ...")
}
