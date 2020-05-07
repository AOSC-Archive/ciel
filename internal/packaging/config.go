package packaging

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path"

	"github.com/AOSC-Dev/ciel/internal/abstract"

	d "github.com/AOSC-Dev/ciel/display"
)

// constant definitions for packaging related variables
const (
	DefaultEditor     = "/usr/bin/editor"
	DefaultRepoConfig = "/etc/apt/sources.list.d/ciel-local.list"
)

// EditSourceList : config function to let user manipulate the apt config inside the container
func EditSourceList(global bool, i abstract.Instance, c abstract.Container) {
	var root string
	if global {
		root = c.DistDir()
	} else {
		root = i.MountPoint()
	}
	editor := editor()
	cmd := exec.Command(editor, path.Join(root, "/etc/apt/sources.list"))
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

// SetTreePath : config function to set the acbs search path
func SetTreePath(global bool, i abstract.Instance, c abstract.Container, tree string) {
	var root string
	if global {
		root = c.DistDir()
	} else {
		root = i.MountPoint()
	}
	config := `[default]` + "\n"
	config += `location = ` + path.Clean(tree) + "\n"
	d.ITEM("set tree path")
	err := ioutil.WriteFile(path.Join(root, "/etc/acbs/forest.conf"), []byte(config), 0644)
	d.ERR(err)
}

// DisableDNSSEC : config function to disable DNSSEC service
func DisableDNSSEC(global bool, i abstract.Instance, c abstract.Container) {
	var root string
	if global {
		root = c.DistDir()
	} else {
		root = i.MountPoint()
	}
	config := `[Resolve]` + "\n"
	config += `DNSSEC=no` + "\n"
	d.ITEM("disable DNSSEC")
	err := ioutil.WriteFile(path.Join(root, "/etc/systemd/resolved.conf"), []byte(config), 0644)
	d.ERR(err)
}

// SetMaintainer : config function to let user specify maintainer information
func SetMaintainer(global bool, i abstract.Instance, c abstract.Container, person string) {
	var root string
	if global {
		root = c.DistDir()
	} else {
		root = i.MountPoint()
	}
	config := `#!/bin/bash` + "\n"
	config += `ABMPM=dpkg` + "\n"
	config += `ABAPMS=` + "\n"
	config += `MTER="` + person + `"` + "\n"
	config += `ABINSTALL=dpkg` + "\n"
	d.ITEM("set maintainer")
	err := ioutil.WriteFile(path.Join(root, "/usr/lib/autobuild3/etc/autobuild/ab3cfg.sh"), []byte(config), 0644)
	d.ERR(err)
}

// InitLocalRepo : initialize local repository
func InitLocalRepo(global bool, i abstract.Instance, c abstract.Container) {
	var root string
	if global {
		root = c.DistDir()
	} else {
		root = i.MountPoint()
	}
	config := `deb file:///debs /` + "\n"
	d.ITEM("initialize local repository")
	err := ioutil.WriteFile(path.Join(root, DefaultRepoConfig), []byte(config), 0644)
	d.ERR(err)
}

// UnInitLocalRepo : remove local repository (configuration only)
func UnInitLocalRepo(global bool, i abstract.Instance, c abstract.Container) {
	var root string
	if global {
		root = c.DistDir()
	} else {
		root = i.MountPoint()
	}
	d.ITEM("un-initialize local repository")
	err := os.Remove(path.Join(root, DefaultRepoConfig))
	d.ERR(err)
}

func editor() string {
	if s := os.Getenv("VISUAL"); s != "" {
		return s
	}
	if s := os.Getenv("EDITOR"); s != "" {
		return s
	}
	return DefaultEditor
}
