package main

// import (
// 	"ciel-driver"
// 	"errors"
// 	"os"
// )
//
// func deploy(c *ciel.Container) error {
// 	c.Fs.Unmount()
// 	os.RemoveAll(c.Fs.TopLayer())
//
// 	if ec := c.Command("apt install -y systemd"); ec != 0 {
// 		return errors.New("apt install systemd: failed")
// 	}
// 	if ec := c.Command("apt install -y admin-base core-base editor-base python-base" +
// 		" network-base systemd-base web-base util-base devel-base debug-base autobuild3 git"); ec != 0 {
// 		return errors.New("apt install {base}: failed")
// 	}
//
// 	c.Fs.Unmount()
// 	c.Fs.DisableAll()
// 	c.Fs.EnableLayer("stub", "buildkit")
// 	if err := cleanRelease(c); err != nil {
// 		return err
// 	}
// 	return c.Fs.MergeFile("/", "upperdir", "buildkit", false)
// }
