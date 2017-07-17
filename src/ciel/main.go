package main

import (
	"ciel-driver"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
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

var subCommand string
var subArgs []string

var cmdTable = map[string]func() int{
	"init":   cielInit,
	"drop":   cielDrop,
	"mount":  cielMount,
	"merge":  cielMerge,
	"clean":  cielClean,
	"shell":  cielShell,
	"rawcmd": cielRawcmd,
	"help":   cielHelp,
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
func main() {
	args := os.Args
	subCommand = DefaultCommand
	subArgs = DefaultCommandArgs
	if len(os.Args) > 1 {
		subCommand = args[1]
		subArgs = args[2:]
	}
	route, exists := cmdTable[subCommand]
	if !exists {
		route = cielPlugin
	}
	os.Exit(route())
}

func requireEUID0() {
	if os.Geteuid() != 0 {
		log.Fatalf("%s: you must be root to run this command\n", subCommand)
	}
}
func requireFS() {
	path, err := filepath.Abs(FileSystemDir)
	if err != nil {
		path = FileSystemDir
	}
	if fi, err := os.Stat(path); os.IsNotExist(err) {
		log.Fatalf("%s: ciel file system %s not found\n", subCommand, path)
	} else if err != nil {
		log.Fatalf("%s: cannot access ciel file system %s: %v\n", subCommand, path, err)
	} else if !fi.IsDir() {
		log.Fatalf("%s: ciel file system %s must be a directory\n", subCommand, path)
	}
}

// ciel init <tarball>
func cielInit() int {
	requireEUID0()
	if len(subArgs) != 1 {
		log.Println("init: you may only input one argument")
		return 1
	}
	err := genesis(subArgs[0], FileSystemDir)
	if err != nil {
		log.Fatalln(err)
	}
	return 0
}

// ciel drop [<layers>]
func cielDrop() int {
	requireEUID0()
	requireFS()
	c := ciel.New(MachineName, FileSystemDir)
	if len(subArgs) == 0 {
		subArgs = []string{"upperdir"}
	}
	for _, layer := range subArgs {
		path := c.Fs.Layer(layer)
		if path == "" {
			log.Printf("drop: layer %s not exist\n", layer)
			continue
		}
		if err := os.RemoveAll(c.Fs.Layer(layer)); err != nil {
			log.Println(err)
		}
	}
	return 0
}

// ciel mount [--read-write] [<layers>]
func cielMount() int {
	requireEUID0()
	requireFS()
	c := ciel.New(MachineName, FileSystemDir)
	var rw = false
	if len(subArgs) >= 1 && subArgs[0] == "--read-write" {
		subArgs = subArgs[1:]
		rw = true
	}
	if len(subArgs) > 0 {
		c.Fs.DisableAll()
		c.Fs.EnableLayer(subArgs...)
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
	return 0
}

// ciel merge [<upper>..]<lower> [--no-self] path
func cielMerge() int {
	requireEUID0()
	requireFS()
	// FIXME: limit arguments
	layers := strings.SplitN(subArgs[0], "..", 2)
	if len(layers) == 1 { // "xx" => ["upperdir" "xx"]
		layers = append([]string{"upperdir"}, layers[0])
	} else if layers[0] == "" { // "..xx" => ["upperdir" "xx"]
		layers[0] = "upperdir"
	}
	var excludeSelf = false
	var path string
	excludeSelf = subArgs[1] == "--no-self"
	if subArgs[1] == "--no-self" {
		path = subArgs[2]
	} else {
		path = subArgs[1]
	}
	c := ciel.New(MachineName, FileSystemDir)
	c.Fs.MergeFile(path, layers[0], layers[1], excludeSelf)
	return 0
}

// ciel clean [--factory-reset]
func cielClean() int {
	requireEUID0()
	requireFS()
	c := ciel.New(MachineName, FileSystemDir)
	c.Fs.DisableLayer("override", "cache")
	var err error
	if len(subArgs) == 1 && subArgs[0] == "--factory-reset" {
		err = cleanFactoryReset(c)
	} else {
		err = cleanNormal(c)
	}
	c.Shutdown()
	c.Fs.Unmount()
	if err != nil {
		log.Fatalln(err)
	}
	return 0
}

// ciel shell [<cmdline>]
func cielShell() int {
	requireEUID0()
	requireFS()
	c := ciel.New(MachineName, FileSystemDir)
	defer c.Fs.Unmount()
	defer c.Shutdown()
	var exitcode int
	if len(subArgs) == 0 {
		exitcode = c.Shell()
	} else if len(subArgs) == 1 {
		exitcode = c.Command(subArgs[0])
	} else {
		log.Println("shell: you may only input one argument")
		return 1
	}
	return exitcode
}

// ciel rawcmd <cmd> <arg1> <arg2> ...
func cielRawcmd() int {
	requireEUID0()
	requireFS()
	if len(subArgs) == 0 {
		log.Println("init: you must input one argument at least")
		return 1
	}
	c := ciel.New(MachineName, FileSystemDir)
	defer c.Fs.Unmount()
	defer c.Shutdown()
	exitcode := c.CommandRaw(subArgs[0], os.Stdin, os.Stdout, os.Stderr, subArgs[1:]...)
	return exitcode
}

func cielHelp() int {
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
	return 0
}

func cielPlugin() int {
	proc := LibExecDir + "/ciel-plugin/ciel-" + subCommand
	cmd := exec.Command(proc, subArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitStatus := exitError.Sys().(syscall.WaitStatus)
			return exitStatus.ExitStatus()
		}
		log.Printf("failed to run plugin %s: %v\n", subCommand, err)
		return 1
	}
	return 0
}
