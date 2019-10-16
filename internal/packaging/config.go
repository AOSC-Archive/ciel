package packaging

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path"

	d "ciel/display"
	"ciel/internal/abstract"
)

const (
	DefaultEditor     = "/usr/bin/editor"
	DefaultRepoConfig = "/etc/apt/sources.list.d/ciel-local.list"
)

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

func editor() string {
	if s := os.Getenv("VISUAL"); s != "" {
		return s
	}
	if s := os.Getenv("EDITOR"); s != "" {
		return s
	}
	return DefaultEditor
}
