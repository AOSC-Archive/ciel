package dotciel

import (
	"io/ioutil"
	"log"
	"path"

	"ciel/internal/dotciel/container"
	"ciel/internal/utils"
	"strings"
)

const (
	DotCielDirName   = ".ciel"
	ContainerDirName = "container"
	VersionFile      = "version"
	Version          = "2"
)

type Ciel struct {
	BasePath string
}

func (i *Ciel) Check() {
	ver, err := ioutil.ReadFile(i.VerFile())
	if err != nil {
		log.Fatalln("not a ciel work directory here")
	}
	if strings.TrimSpace(string(ver)) != Version {
		log.Fatalln("your cieldir is an incompatible version")
	}
}
func (i *Ciel) CielDir() string {
	return path.Join(i.BasePath, DotCielDirName)
}
func (i *Ciel) VerFile() string {
	return path.Join(i.CielDir(), VersionFile)
}

func (i *Ciel) containerDir() string {
	return path.Join(i.CielDir(), ContainerDirName)
}

func (i *Ciel) Init() {
	utils.Mkdir(i.CielDir())
	if err := ioutil.WriteFile(i.VerFile(), []byte(Version), 0644); err != nil {
		log.Panic(err)
	}
	i.Container().Init()
}

func (i *Ciel) Container() *container.Container {
	return &container.Container{Parent: i, BasePath: i.containerDir()}
}
func (i *Ciel) GetBasePath() string { return i.BasePath }
