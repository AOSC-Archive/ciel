package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

const (
	PluginDir    = LibExecDir + "/ciel-plugin"
	PluginPrefix = "ciel-"
)

// Plugin defines a ciel plugin
type Plugin struct {
	Name string
	// TODO: get script usage from header comments
	Usage string
}

func cielPlugin() int {
	proc := filepath.Join(PluginDir, PluginPrefix+SubCommand)
	cmd := exec.Command(proc, SubArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitStatus := exitError.Sys().(syscall.WaitStatus)
			return exitStatus.ExitStatus()
		}
		log.Printf("failed to run plugin %s: %v\n", SubCommand, err)
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
