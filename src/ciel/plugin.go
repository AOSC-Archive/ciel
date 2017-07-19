package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

const (
	PluginDir = LibExecDir + "/ciel-plugin/"
)

// Plugin defines a ciel plugin
type Plugin struct {
	Name string
	// TODO: get script usage from header comments
	Usage string
}

func cielPlugin() int {
	proc := PluginDir + "ciel-" + SubCommand
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
	err := filepath.Walk(PluginDir, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if f.IsDir() {
			return nil
		}
		fname := f.Name()
		if len(fname) > 5 && fname[:5] == `ciel-` {
			Plugins = append(Plugins, Plugin{Name: fname[5:]})
		}
		return nil
	})
	if err != nil {
		log.Fatalf("failed to walk in directory: %v\n", err)
	}
	return Plugins
}
