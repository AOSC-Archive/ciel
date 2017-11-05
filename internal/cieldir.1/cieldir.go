package cieldir

import (
	"io/ioutil"
	"log"
	"path"

	"ciel/internal/utils"
)

const (
	DotCielDirName = ".ciel"
	DistDirName    = "dist"
	InstDirName    = "instances"
	VersionFile    = "version"
	Version        = "1"
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
		log.Fatalln("your cieldir is an incompatible newer version")
	}
}
func (i *CielDir) CielDir() string {
	return path.Join(i.BasePath, DotCielDirName)
}
func (i *CielDir) DistDir() string {
	return path.Join(i.CielDir(), DistDirName)
}
func (i *CielDir) InstDir() string {
	return path.Join(i.CielDir(), InstDirName)
}
func (i *CielDir) VerFile() string {
	return path.Join(i.CielDir(), VersionFile)
}

func (i *CielDir) Init() {
	utils.Mkdir(i.CielDir())
	utils.Mkdir(i.DistDir())
	utils.Mkdir(i.InstDir())
	if err := ioutil.WriteFile(i.VerFile(), []byte(Version), 0644); err != nil {
		log.Panic(err)
	}
}
