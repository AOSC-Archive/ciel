package container

import (
	"log"
	"os"
	"path"

	"ciel/internal/ciel/abstract"
	"ciel/internal/ciel/container/instance"
	"ciel/internal/utils"
)

const (
	DistDirName = "dist"
	InstDirName = "instances"
)

type Container struct {
	Parent   abstract.Ciel
	BasePath string
}

func (i *Container) DistDir() string {
	return path.Join(i.BasePath, DistDirName)
}
func (i *Container) InstDir() string {
	return path.Join(i.BasePath, InstDirName)
}

func (i *Container) Init() {
	utils.Mkdir(i.BasePath)
	utils.Mkdir(i.DistDir())
	utils.Mkdir(i.InstDir())
}

func (i *Container) Instance(name string) *instance.Instance {
	return &instance.Instance{Parent: i, BasePath: i.InstDir(), Name: name}
}

func (i *Container) AddInst(name string) error {
	utils.Mkdir(path.Join(i.InstDir(), name))
	return i.Instance(name).Init()
}
func (i *Container) DelInst(name string) error {
	return os.RemoveAll(path.Join(i.InstDir(), name))
}

func (i *Container) InstExists(name string) bool {
	instDir := path.Join(i.InstDir(), name)
	if instDir == path.Clean(i.InstDir()) {
		return false
	}
	if stat, err := os.Stat(instDir); err != nil || !stat.IsDir() {
		return false
	}
	return true
}
func (i *Container) CheckInst(name string) {
	if !i.InstExists(name) {
		log.Fatalln("instance '" + name + "' does not exist")
	}
}

func (i *Container) GetAll() []*instance.Instance {
	list := i.GetAllNames()
	var instList []*instance.Instance
	for _, name := range list {
		instList = append(instList, i.Instance(name))
	}
	return instList
}
func (i *Container) GetAllNames() []string {
	dir, err := os.Open(i.InstDir())
	if err != nil {
		log.Panic(err)
	}
	subDirs, err := dir.Readdir(0)
	if err != nil {
		log.Panic(err)
	}
	var subDirNames []string
	for _, subDirs := range subDirs {
		if subDirs.IsDir() {
			subDirNames = append(subDirNames, subDirs.Name())
		}
	}
	return subDirNames
}

func (i *Container) GetBasePath() string    { return i.BasePath }
func (i *Container) GetCiel() abstract.Ciel { return i.Parent }
