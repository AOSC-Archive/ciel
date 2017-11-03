package cieldir

import (
	"os"
	"path"
	"log"
	"errors"

	"ciel/internal/overlayfs"
)

const LayerDirName = "layers"

var Layers = []string{
	"local",
	"diff",
}

var (
	ErrLock = errors.New("failed to require the lock")
)

func (i *CielDir) CreateInst(name string) {
	mkdir(i.InstSubDir(name))
	layersDir := path.Join(i.InstSubDir(name), LayerDirName)
	mkdir(layersDir)
	for _, dir := range Layers {
		mkdir(path.Join(layersDir, dir))
	}
}

func (i *CielDir) DeleteInst(name string) {
	os.RemoveAll(i.InstSubDir(name))
}

func (i *CielDir) HasInst(name string) bool {
	if stat, err := os.Stat(i.InstSubDir(name)); err != nil || !stat.IsDir() {
		return false
	}
	return true
}

func (i *CielDir) InstFileSystem(name string) overlayfs.Instance {
	layersDir := path.Join(i.InstSubDir(name), LayerDirName)
	var layers = []string{i.DistDir()}
	for _, dir := range Layers {
		layers = append(layers, path.Join(layersDir, dir))
	}
	return overlayfs.Instance{
		MountPoint: "./" + name,
		Layers:     layers,
	}
}

func (i *CielDir) MountInst(name string) error {
	ofs := i.InstFileSystem(name)
	if !lock(i.InstLock(name)) {
		return ErrLock
	}
	if err := ofs.Mount(false); err != nil {
		unlock(i.InstLock(name))
		return err
	}
	return nil
}

func (i *CielDir) UnmountInst(name string) error {
	ofs := i.InstFileSystem(name)
	if err := ofs.Unmount(); err != nil {
		return err
	}
	os.Remove(ofs.MountPoint)
	unlock(i.InstLock(name))
	return nil
}

func (i *CielDir) InstStat(name string) string {
	if locked(i.InstLock(name)) {
		return "mounted"
	}
	return "free"
}

func (i *CielDir) InstLock(name string) string {
	return path.Join(i.InstSubDir(name), "lock")
}

func (i *CielDir) AllInst() []string {
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

func (i *CielDir) MountAll() {
	instList := i.AllInst()
	for _, inst := range instList {
		err := i.MountInst(inst)
		if err != nil {
			log.Println(inst+":", err)
		} else {
			log.Println(inst+":", "done")
		}
	}
}

func (i *CielDir) UnmountAll() {
	instList := i.AllInst()
	for _, inst := range instList {
		err := i.UnmountInst(inst)
		if err != nil {
			log.Println(inst+":", err)
		} else {
			log.Println(inst+":", "done")
		}
	}
}
