package cieldir

import (
	"io/ioutil"
	"log"
	"path"
)

const (
	CielDirName = ".ciel"
	DistDirName = "dist"
	InstDirName = "instances"
	VersionFile = "version"
	Version     = "1"
)

type CielDir struct {
	BasePath string
}

func (i *CielDir) Check() {
	ver, err := ioutil.ReadFile(i.VerFile())
	if err != nil {
		return
	}
	if string(ver) != Version {
		log.Panicln("your cieldir is an incompatible newer version")
	}
}
func (i *CielDir) CielDir() string {
	return path.Join(i.BasePath, CielDirName)
}
func (i *CielDir) DistDir() string {
	return path.Join(i.CielDir(), DistDirName)
}
func (i *CielDir) InstDir() string {
	return path.Join(i.CielDir(), InstDirName)
}
func (i *CielDir) InstSubDir(name string) string {
	return path.Join(i.InstDir(), name)
}
func (i *CielDir) VerFile() string {
	return path.Join(i.CielDir(), VersionFile)
}

func (i *CielDir) Init() {
	mkdir(i.CielDir())
	mkdir(i.DistDir())
	mkdir(i.InstDir())
	if err := ioutil.WriteFile(i.VerFile(), []byte(Version), 0644); err != nil {
		log.Panic(err)
	}
}
