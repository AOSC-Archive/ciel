package main

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"ciel-driver"
)

const (
	MachineName   = "ciel"
	FileSystemDir = "./cielfs"
	LibExecDir    = "/usr/libexec"
	EnvLogLevel   = "CIEL_LOGLEVEL"
)

var (
	DefaultCommand     = "shell"
	DefaultCommandArgs = []string{}
)

var SubCommand string
var SubArgs []string

var CmdTable = map[string]func() int{
	"init":    cielInit,
	"drop":    cielDrop,
	"mount":   cielMount,
	"merge":   cielMerge,
	"clean":   cielClean,
	"shell":   cielShell,
	"rawcmd":  cielRawcmd,
	"help":    cielHelp,
	"version": cielVersion,
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
	logLevel, _ := strconv.Atoi(os.Getenv(EnvLogLevel))
	SetLogLevel(logLevel)
	ciel.SetLogLevel(logLevel)
}
func main() {
	args := os.Args
	SubCommand = DefaultCommand
	SubArgs = DefaultCommandArgs
	if len(os.Args) > 1 {
		SubCommand = args[1]
		SubArgs = args[2:]
	}
	route, exists := CmdTable[SubCommand]
	if !exists {
		route = cielPlugin
	}
	os.Exit(route())
}

func requireEUID0() {
	if os.Geteuid() != 0 {
		errlog.Fatalf("%s: need to be root\n", SubCommand)
	}
}
func requireFS() {
	path, err := filepath.Abs(FileSystemDir)
	if err != nil {
		path = FileSystemDir
	}
	if fi, err := os.Stat(path); os.IsNotExist(err) {
		errlog.Fatalf("%s: ciel file system %s not found\n", SubCommand, path)
	} else if err != nil {
		errlog.Fatalf("%s: cannot access ciel file system %s: %v\n", SubCommand, path, err)
	} else if !fi.IsDir() {
		errlog.Fatalf("%s: ciel file system %s must be a directory\n", SubCommand, path)
	}
}

// ciel init <tarball>
func cielInit() int {
	requireEUID0()
	if len(SubArgs) != 1 {
		errlog.Println("init: you may only input one argument")
		return 1
	}
	err := genesis(SubArgs[0], FileSystemDir)
	if err != nil {
		errlog.Println(err)
		return 1
	}
	return 0
}

// ciel drop [<layers>]
func cielDrop() int {
	requireEUID0()
	requireFS()
	c := ciel.New(MachineName, FileSystemDir)
	if len(SubArgs) == 0 {
		SubArgs = []string{"upperdir"}
	}
	var err error
	for _, layer := range SubArgs {
		if err = os.RemoveAll(c.Fs.Layer(layer)); err != nil {
			warnlog.Println(err)
		}
	}
	if err != nil {
		return 1
	}
	return 0
}

// ciel mount [--read-write] [<layers>]
func cielMount() int {
	requireEUID0()
	requireFS()
	c := ciel.New(MachineName, FileSystemDir)
	var rw = false
	if len(SubArgs) >= 1 && SubArgs[0] == "--read-write" {
		SubArgs = SubArgs[1:]
		rw = true
	}
	if len(SubArgs) > 0 {
		c.Fs.DisableAll()
		c.Fs.EnableLayer(SubArgs...)
	}
	var err error
	if rw {
		err = c.Fs.Mount()
	} else {
		err = c.Fs.MountReadOnly()
	}
	if err != nil {
		errlog.Println(err)
		return 1
	}
	fmt.Println(c.Fs.TargetDir())
	return 0
}

// ciel merge [<upper>..]<lower> [--no-self] path
func cielMerge() int {
	requireEUID0()
	requireFS()
	// FIXME: limit arguments
	layers := strings.SplitN(SubArgs[0], "..", 2)
	if len(layers) == 1 { // "xx" => ["upperdir" "xx"]
		layers = append([]string{"upperdir"}, layers[0])
	} else if layers[0] == "" { // "..xx" => ["upperdir" "xx"]
		layers[0] = "upperdir"
	}
	var excludeSelf = false
	var path string
	excludeSelf = SubArgs[1] == "--no-self"
	if SubArgs[1] == "--no-self" {
		path = SubArgs[2]
	} else {
		path = SubArgs[1]
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
	if len(SubArgs) == 1 && SubArgs[0] == "--factory-reset" {
		err = cleanFactoryReset(c)
	} else {
		err = cleanNormal(c)
	}
	c.Shutdown()
	c.Fs.Unmount()
	if err != nil {
		errlog.Println(err)
		return 1
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
	var exitStatus int
	if len(SubArgs) == 0 {
		exitStatus = c.Shell()
	} else if len(SubArgs) == 1 {
		exitStatus = c.Command(SubArgs[0])
	} else {
		errlog.Println("shell: you may only input one argument")
		return 1
	}
	return exitStatus
}

// ciel rawcmd <cmd> <arg1> <arg2> ...
func cielRawcmd() int {
	requireEUID0()
	requireFS()
	if len(SubArgs) == 0 {
		errlog.Println("rawcmd: you must input one argument at least")
		return 1
	}
	c := ciel.New(MachineName, FileSystemDir)
	defer c.Fs.Unmount()
	defer c.Shutdown()
	exitStatus := c.CommandRaw(SubArgs[0], os.Stdin, os.Stdout, os.Stderr, SubArgs[1:]...)
	return exitStatus
}

func cielHelp() int {
	fmt.Println("Usage: " + os.Args[0] + " [command [...]]")
	fmt.Println(`Default command is "shell".`)
	fmt.Println("")

	fmt.Println("Commands:")
	fmt.Println("\thelp")
	fmt.Println("\tversion")
	fmt.Println("")
	fmt.Println("\tinit   <tarball>")
	fmt.Println("")
	fmt.Println("\tdrop   [<layers>]")
	fmt.Println("\tmount  [--read-write] [<layers>]")
	fmt.Println("\tmerge  [<upper>..]<lower> [--no-self] path")
	fmt.Println("\tclean  [--factory-reset]")
	fmt.Println("")
	fmt.Println("\tshell  [<cmdline>]")
	fmt.Println("\trawcmd <cmd> <arg1> <arg2> ...")
	fmt.Println("")

	fmt.Println("Plugins:")
	Plugins := getPlugins()
	for _, Plugin := range Plugins {
		fmt.Printf("\t%s\n", Plugin.Name)
	}
	return 0
}

func cielVersion() int {
	fmt.Println(Version)
	return 0
}
