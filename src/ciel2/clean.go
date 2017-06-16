package main

import (
	"ciel-driver"
	"errors"
	"os"
	"path/filepath"
	"regexp"
)

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
	c.Mount()
	fs := c.FileSystem()
	if fs.Target == "" {
		return errors.New("mount file system first")
	}
	filepath.Walk(fs.Target, wrapWalkFunc(fs.Target, func(path string, info os.FileInfo, err error) error {
		if _, indpkg := dpkgfiles[path]; indpkg {
			return nil
		}
		if !(regex.MatchString(path) && !regexN.MatchString(path)) {
			return fn(filepath.Join(fs.Target, path), info, err)
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
