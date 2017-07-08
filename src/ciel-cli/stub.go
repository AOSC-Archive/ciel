package main

import (
	"ciel-driver"
	"os"
)

func updateStub(c *ciel.Container) error {
	defer func() {
		c.Fs.EnableAll()
		c.Fs.Unmount()
	}()

	c.Fs.Unmount()
	os.RemoveAll(c.Fs.TopLayer())
	os.RemoveAll(c.Fs.Layer("dist"))
	c.Fs.DisableAll()
	c.Fs.EnableLayer("stub", "stub-overlay")

	if err := aptUpdate(c); err != nil {
		return err
	}

	c.Fs.Unmount()
	c.Fs.DisableAll()
	c.Fs.EnableLayer("stub")
	if err := cleanRelease(c); err != nil {
		return err
	}
	return c.Fs.MergeFile("/", "upperdir", "stub", false)
}
