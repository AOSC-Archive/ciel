package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

//...
// TODO: plugin system

const (
	LibExecDir   = "/usr/libexec"
	PluginDir    = LibExecDir + "/ciel-plugin"
	PluginPrefix = "ciel-"
)

// Plugin defines a ciel plugin
type Plugin struct {
	Name string
	// TODO: get script usage from header comments
	Usage string
}

func plugin(subCmd string) int {
	basePath := flagCielDir()
	instName := flagInstance()
	networkFlag := flagNetwork()
	noBooting := flagNoBooting()
	bootConfig := flagBootConfig()
	parse()
	saveCielDir(*basePath)
	saveInstance(*instName)
	saveNetwork(*networkFlag)
	saveNoBooting(*noBooting)
	saveBootConfig(*bootConfig)

	proc := filepath.Join(PluginDir, PluginPrefix+subCmd)
	cmd := exec.Command(proc, flag.Args()...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// err := cmd.Run()

	println("script:", proc, strings.Join(flag.Args(), " "))
	println("not implemented")
	var err error

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitStatus := exitError.Sys().(syscall.WaitStatus)
			return exitStatus.ExitStatus()
		}
		log.Printf("failed to run plugin %s: %v\n", subCmd, err)
		return 1
	}
	return 0
}

func getPlugins() []Plugin {
	var Plugins []Plugin
	files, err := ioutil.ReadDir(PluginDir)
	if err != nil {
		log.Fatalf("failed to get files under plugin directory: %v\n", err)
	}
	for _, f := range files {
		if f.IsDir() {
			continue
		} else {
			fname := f.Name()
			if len(fname) > len(PluginPrefix) && strings.HasPrefix(fname, PluginPrefix) {
				Plugins = append(Plugins, Plugin{Name: fname[len(PluginPrefix):]})
			}
		}
	}
	return Plugins
}
