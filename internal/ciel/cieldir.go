package ciel

import (
	"io/ioutil"
	"log"
	"path"
	"strings"

	"ciel/internal/abstract"
	"ciel/internal/container"
	"ciel/internal/packaging"
	"ciel/internal/pkgtree"
	"ciel/internal/utils"
)

const (
	DotCielDirName = ".ciel"

	ContainerDirName = DotCielDirName + "/container"
	TreeDirName      = "TREE"
	OutputDirName    = "OUTPUT/debs"

	VersionFile = DotCielDirName + "/version"
	Version     = "2"
)

type Ciel struct {
	BasePath string
}

func (i *Ciel) Check() {
	ver, err := ioutil.ReadFile(i.VerFile())
	if err != nil {
		log.Fatalln("not a Ciel work directory here")
	}
	if strings.TrimSpace(string(ver)) != Version {
		log.Fatalln("your Ciel work directory is an incompatible version")
	}
}
func (i *Ciel) CielDir() string {
	return path.Join(i.BasePath, DotCielDirName)
}
func (i *Ciel) VerFile() string {
	return path.Join(i.BasePath, VersionFile)
}
func (i *Ciel) containerDir() string {
	return path.Join(i.BasePath, ContainerDirName)
}
func (i *Ciel) treeDir() string {
	return path.Join(i.BasePath, TreeDirName)
}
func (i *Ciel) outDir() string {
	return path.Join(i.BasePath, OutputDirName)
}

func (i *Ciel) Init() {
	utils.MustMkdir(i.CielDir())
	if err := ioutil.WriteFile(i.VerFile(), []byte(Version), 0644); err != nil {
		log.Panic(err)
	}
	i.Container().Init()
}

func (i *Ciel) Container() *container.Container {
	return &container.Container{Parent: i, BasePath: i.containerDir()}
}
func (i *Ciel) Tree() *pkgtree.Tree {
	return &pkgtree.Tree{Parent: i, BasePath: i.treeDir()}
}
func (i *Ciel) Output() *packaging.Tree {
	return &packaging.Tree{Parent: i, BasePath: i.outDir()}
}
func (i *Ciel) GetContainer() abstract.Container {
	return i.Container()
}
func (i *Ciel) GetTree() abstract.Tree {
	return i.Tree()
}
func (i *Ciel) GetOutput() abstract.Tree {
	return i.Output()
}
func (i *Ciel) GetBasePath() string { return i.BasePath }
