package ciel

import (
	"encoding/base64"
	"math/rand"
	"os"
	"reflect"
	"strings"
	"syscall"
)

type filesystem struct {
	workdir    string `role:"work"  dir:"99-workdir"`
	upperdir   string `role:"upper" dir:"99-upperdir"`
	cache      string `role:"lower" dir:"50-cache"`
	buildkit   string `role:"lower" dir:"10-buildkit"`
	stubConfig string `role:"lower" dir:"01-stub-config"`
	stub       string `role:"lower" dir:"00-stub"`
	base       string
	target     string
	active     bool
}

const _SYSTEMDPATH = "/usr/lib/systemd/systemd"

func (fs *filesystem) isBootable() bool {
	if !fs.active {
		return false
	}
	if _, err := os.Stat(fs.target + _SYSTEMDPATH); os.IsNotExist(err) {
		return false
	}
	return true
}
func (fs *filesystem) setBaseDir(path string) {
	fs.base = path
	t := reflect.TypeOf(*fs)
	v := reflect.ValueOf(fs).Elem()
	n := t.NumField()
	for i := 0; i < n; i++ {
		role := t.Field(i).Tag.Get("role")
		dir := t.Field(i).Tag.Get("dir")
		if dir != "" {
			fulldir := fs.base + "/" + dir
			v.Field(i).SetString(fulldir)
			if role != "work" {
				os.Mkdir(fulldir, 0775)
			}
		}
	}
}
func randomFilename() string {
	const SIZE = 8
	rd := make([]byte, SIZE)
	if _, err := rand.Read(rd); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(rd)
}
func (fs *filesystem) mount() error {
	// NOTE: reflect c.fs, create all directories, stack lowerdirs into argument, mount.
	lowerdirs := []string{}
	t := reflect.TypeOf(*fs)
	v := reflect.ValueOf(fs).Elem()
	n := t.NumField()
	for i := 0; i < n; i++ {
		role := t.Field(i).Tag.Get("role")
		if role == "lower" {
			lowerdirs = append(lowerdirs, v.Field(i).String())
		}
	}
	fs.target = "/tmp/ciel." + randomFilename()
	return mount(fs.target, fs.upperdir, fs.workdir, lowerdirs...)
}
func (fs *filesystem) unmount() error {
	if err := unmount(fs.target); err != nil {
		return err
	}
	if err := os.Remove(fs.target); err != nil {
		return err
	}
	if err := os.RemoveAll(fs.workdir); err != nil {
		return err
	}
	return nil
}
func mount(path string, upperdir string, workdir string, lowerdirs ...string) error {
	return syscall.Mount("overlay", path, "overlay", 0,
		"lowerdir="+strings.Join(lowerdirs, ":")+",upperdir="+upperdir+",workdir="+workdir)
}
func unmount(path string) error {
	return syscall.Unmount(path, 0)
}
