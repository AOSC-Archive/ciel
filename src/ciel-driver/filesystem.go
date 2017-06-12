package ciel

import (
	"encoding/base64"
	"math/rand"
	"os"
	"reflect"
	"strings"
	"sync"
	"syscall"
)

type filesystem struct {
	lock sync.RWMutex

	Workdir    string `role:"work"  dir:"99-workdir"`
	Upperdir   string `role:"upper" dir:"99-upperdir"`
	Cache      string `role:"lower" dir:"50-cache"`
	Buildkit   string `role:"lower" dir:"10-buildkit"`
	StubConfig string `role:"lower" dir:"01-stub-config"`
	Stub       string `role:"lower" dir:"00-stub"`

	base   string
	target string
	active bool
}

const _SYSTEMDPATH = "/usr/lib/systemd/systemd"

func (fs *filesystem) IsBootable() bool {
	fs.lock.RLock()
	defer fs.lock.RUnlock()

	if !fs.active {
		return false
	}
	if _, err := os.Stat(fs.target + _SYSTEMDPATH); os.IsNotExist(err) {
		return false
	}
	return true
}

func (fs *filesystem) IsActive() bool {
	fs.lock.RLock()
	defer fs.lock.RUnlock()
	return fs.active
}

func (fs *filesystem) SetBaseDir(path string) {
	fs.lock.Lock()
	defer fs.lock.Unlock()

	if fs.active {
		panic("setBaseDir when filesystem is active")
	}

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

func (fs *filesystem) Mount() error {
	fs.lock.Lock()
	defer fs.lock.Unlock()

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
	os.Mkdir(fs.target, 0775)
	os.Mkdir(fs.Workdir, 0775)
	reterr := mount(fs.target, fs.Upperdir, fs.Workdir, lowerdirs...)
	if reterr == nil {
		fs.active = true
	}
	return reterr
}

func (fs *filesystem) Unmount() error {
	fs.lock.Lock()
	defer fs.lock.Unlock()
	if !fs.active {
		return nil
	}

	if err := unmount(fs.target); err != nil {
		return err
	}
	defer func() {
		fs.active = false
	}()
	err1 := os.Remove(fs.target)
	err2 := os.RemoveAll(fs.Workdir)
	if err2 != nil {
		return err2
	}
	if err1 != nil {
		return err1
	}
	return nil
}

func randomFilename() string {
	const SIZE = 8
	rd := make([]byte, SIZE)
	if _, err := rand.Read(rd); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(rd)
}
func mount(path string, upperdir string, workdir string, lowerdirs ...string) error {
	return syscall.Mount("overlay", path, "overlay", 0,
		"lowerdir="+strings.Join(lowerdirs, ":")+",upperdir="+upperdir+",workdir="+workdir)
}
func unmount(path string) error {
	return syscall.Unmount(path, 0)
}
