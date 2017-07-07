package main

import (
	"ciel-driver"
	"errors"
	"os"
	"os/exec"
)

func gensis(tar, fsroot string) error {
	os.Mkdir(fsroot, 0755)
	c := ciel.New("temp", fsroot)
	if err := c.Fs.BuildDirs(); err != nil {
		return err
	}
	cmd := exec.Command("/bin/tar", "-xf", tar, "-C", c.Fs.Layer("stub"))
	b, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New(string(b))
	}
	return nil
}
