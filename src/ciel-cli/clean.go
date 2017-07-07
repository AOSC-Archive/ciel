package main

import (
	"ciel-driver"
	"errors"
	"os"
	"path/filepath"
	"regexp"
)

func cleanRelease(c *ciel.Container) error {
	return clean(c, []string{
		`^/dev`,
		`^/efi`,
		`^/etc`,
		`^/run`,
		`^/usr`,
		`^/var/lib/dpkg`,
		`^/var/log/journal$`,
		`^/root`,
		`^/home/aosc`,
		`/\.updated$`,
	}, []string{
		`^/etc/.*-$`,
		`^/etc/machine-id`,
		`^/etc/ssh/ssh_host_.*`,
		`^/var/lib/dpkg/status-old`,
	}, func(path string, info os.FileInfo, err error) error {
		if err := os.RemoveAll(path); err != nil {
			println(path, err.Error())
		}
		return nil
	})
}

func cleanBoot(c *ciel.Container) error {
	// TODO
	return nil
}

func clean(c *ciel.Container, re []string, reN []string, fn filepath.WalkFunc) error {
	restr := "((" + re[0] + ")"
	for _, re := range re[1:] {
		restr += "|(" + re + ")"
	}
	restr += ")"
	restrN := "((" + reN[0] + ")"
	for _, re := range reN[1:] {
		restrN += "|(" + re + ")"
	}
	restrN += ")"
	regex := regexp.MustCompile(restr)
	regexN := regexp.MustCompile(restrN)
	dpkgfiles := dpkgPackageFiles(c)
	if dpkgfiles == nil {
		return errors.New("no file in dpkg")
	}
	c.Shutdown()
	c.Fs.Mount()
	if !c.Fs.IsMounted() {
		return errors.New("cannot mount file system")
	}
	target := c.Fs.TargetDir()
	filepath.Walk(target, wrapWalkFunc(target, func(path string, info os.FileInfo, err error) error {
		if _, indpkg := dpkgfiles[path]; indpkg {
			return nil
		}
		if !(regex.MatchString(path) && !regexN.MatchString(path)) {
			return fn(filepath.Join(target, path), info, err)
		} else {
			//println(path)
		}
		return nil
	}))

	return nil
}

func wrapWalkFunc(root string, fn filepath.WalkFunc) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if info == nil {
			return err
		}
		if path == root {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		rel = "/" + rel
		return fn(rel, info, err)
	}
}
