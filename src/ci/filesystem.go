package ci

import (
	"encoding/base64"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

type ContainerFilesystem struct {
	Base string

	Stub         string
	StubConfig   string
	Buildkit     string
	Cache        string
	Upperdir     string
	UpperdirWork string

	Target string
}

func InitFilesystem(bkdir string) *ContainerFilesystem {
	fs := &ContainerFilesystem{Base: bkdir}
	fs.Stub = fs.Base + "/00-stub"
	fs.StubConfig = fs.Base + "/01-stub"
	fs.Buildkit = fs.Base + "/10-buildkit"
	fs.Cache = fs.Base + "/50-cache"
	fs.Upperdir = fs.Base + "/99-upperdir"
	fs.UpperdirWork = fs.Base + "/99-upperdir-work"
	os.Mkdir(fs.Base, 0755)
	os.Mkdir(fs.Stub, 0755)
	os.Mkdir(fs.StubConfig, 0755)
	os.Mkdir(fs.Buildkit, 0755)
	os.Mkdir(fs.Cache, 0755)
	os.Mkdir(fs.Upperdir, 0755)
	return fs
}

func randomFilename(size int) string {
	rd := make([]byte, size)
	if _, err := rand.Read(rd); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(rd)
}

func (fs *ContainerFilesystem) Startup() error {
	os.Mkdir(fs.UpperdirWork, 0755)
	fs.Target = os.TempDir() + "/ciel." + randomFilename(8)
	os.Mkdir(fs.Target, 0755)
	return mount(fs.Target, fs.Upperdir, fs.UpperdirWork,
		fs.Cache,
		fs.Buildkit,
		fs.StubConfig,
		fs.Stub,
	)
}

func (fs *ContainerFilesystem) Shutdown() error {
	if err := unmount(fs.Target); err != nil {
		return err
	}
	if err := os.Remove(fs.Target); err != nil {
		return err
	}
	if err := os.RemoveAll(fs.UpperdirWork); err != nil {
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

func (fs *ContainerFilesystem) UpperDir(path string) string {
	return fs.Upperdir + path
}

func (fs *ContainerFilesystem) Merge(path string, targetLayer string) error {
	err := filepath.Walk(fs.UpperDir(path), func(path string, info os.FileInfo, err error) error {
		if info == nil {
			return err
		}
		rel, err := filepath.Rel(fs.UpperDir("/"), path)
		if err != nil {
			return err
		}
		rel = "/" + rel
		if info.IsDir() {
			os.MkdirAll(targetLayer+rel, 755)
		}
		if err := os.Rename(path, targetLayer+rel); err == nil {
			log.Println("clean: merge", rel)
		}
		return nil
	})
	return err
}
