package main

import (
	"ciel-driver"
	// "encoding/json"
	"errors"
	"os"
	"strings"
)

func configDist(c *ciel.Container) error {
	c.Fs.Unmount()
	os.RemoveAll(c.Fs.TopLayer())
	c.Fs.DisableAll()
	c.Fs.EnableLayer("stub", "stub-overlay", "dist", "dist-overlay")

	var packages = []string{
		"systemd",
	}
	if ec := c.Command("apt install -y " + strings.Join(packages, " ")); ec != 0 {
		return errors.New("apt install systemd: failed")
	}

	packages = []string{
		"admin-base", "core-base", "editor-base", "python-base",
		"network-base", "systemd-base", "web-base", "util-base",
		"devel-base", "debug-base", "autobuild3", "git",
	}

	if ec := c.Command("apt install -y " + strings.Join(packages, " ")); ec != 0 {
		return errors.New("apt install {base}: failed")
	}

	c.Fs.Unmount()
	c.Fs.DisableAll()
	c.Fs.EnableLayer("stub", "stub-overlay", "dist")
	if err := cleanRelease(c); err != nil {
		return err
	}
	return c.Fs.MergeFile("/", "upperdir", "dist", false)
}

func updateDist(c *ciel.Container) error {
	defer func() {
		c.Fs.EnableAll()
		c.Fs.Unmount()
	}()

	c.Fs.Unmount()
	os.RemoveAll(c.Fs.TopLayer())
	c.Fs.DisableAll()
	c.Fs.EnableLayer("stub", "stub-overlay", "dist", "dist-overlay")

	if err := aptUpdate(c); err != nil {
		return err
	}

	c.Fs.Unmount()
	c.Fs.DisableAll()
	c.Fs.EnableLayer("stub", "stub-overlay", "dist")
	if err := cleanRelease(c); err != nil {
		return err
	}
	return c.Fs.MergeFile("/", "upperdir", "dist", false)
}
