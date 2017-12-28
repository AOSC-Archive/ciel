package packaging

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path"

	"ciel/internal/ciel/abstract"
	"ciel/internal/display"
	"ciel/internal/utils"
)

func EditSourceList(i abstract.Instance) {
	root := i.MountPoint()
	editor := utils.Editor()
	cmd := exec.Command(editor, path.Join(root, "/etc/apt/sources.list"))
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func SetTreePath(i abstract.Instance, tree string) {
	root := i.MountPoint()
	config := `[default]` + "\n"
	config += `location = ` + path.Clean(tree) + "\n"
	d.ITEM("set tree path")
	err := ioutil.WriteFile(path.Join(root, "/etc/acbs/forest.conf"), []byte(config), 0644)
	d.ERR(err)
}

func SetMaintainer(i abstract.Instance, person string) {
	root := i.MountPoint()
	config := `#!/bin/bash` + "\n"
	config += `ABMPM=dpkg` + "\n"
	config += `ABAPMS=` + "\n"
	config += `MTER="` + person + `"` + "\n"
	config += `ABINSTALL=` + "\n"
	d.ITEM("set maintainer")
	err := ioutil.WriteFile(path.Join(root, "/usr/lib/autobuild3/etc/autobuild/ab3cfg.sh"), []byte(config), 0644)
	d.ERR(err)
}
