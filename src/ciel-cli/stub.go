package main

import (
	"ciel-driver"
	"errors"
	"os"
)

func updateStub(c *ciel.Container) error {
	defer func() {
		c.Fs.EnableAll()
		c.Fs.Unmount()
	}()

	c.Fs.Unmount()
	os.RemoveAll(c.Fs.TopLayer())
	os.RemoveAll(c.Fs.Layer("buildkit"))
	c.Fs.DisableAll()
	c.Fs.EnableLayer("stub", "stub-config")

	if ec := c.Command("apt update -y"); ec != 0 {
		return errors.New("apt update: failed")
	}
	if ec := c.Command("apt full-upgrade -y"); ec != 0 {
		return errors.New("apt full-upgrade: failed")
	}

	c.Fs.Unmount()
	c.Fs.DisableAll()
	c.Fs.EnableLayer("stub")
	if err := cleanRelease(c); err != nil {
		return err
	}
	return c.Fs.MergeFile("/", "upperdir", "stub", false)
}
