package container

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	"ciel/internal/abstract"
	"ciel/internal/container/instance"
	"ciel/internal/utils"
)

const (
	DistDirName = "dist"
	InstDirName = "instances"
)

var (
	ErrInvalidInstName = errors.New("invalid instance name")
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
	utils.MustMkdir(i.BasePath)
	utils.MustMkdir(i.DistDir())
	utils.MustMkdir(i.InstDir())
}

func (i *Container) Instance(name string) *instance.Instance {
	return &instance.Instance{Parent: i, BasePath: i.InstDir(), Name: name}
}

func (i *Container) AddInst(name string) error {
	if strings.ContainsAny(name, "/\\ ") {
		return ErrInvalidInstName
	}
	utils.MustMkdir(path.Join(i.InstDir(), name))
	return i.Instance(name).Init()
}
func (i *Container) DelInst(name string) error {
	i.Instance(name).RunLock().Remove()
	i.Instance(name).FileSystemLock().Remove()
	return os.RemoveAll(path.Join(i.InstDir(), name))
}

func (i *Container) InstExists(name string) bool {
	if strings.ContainsAny(name, "/\\ ") {
		return false
	}
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
	subDirs, err := ioutil.ReadDir(i.InstDir())
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
