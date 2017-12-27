package main

import (
	"context"
	"os"

	"ciel/internal/ciel"
	"ciel/internal/ciel/container/instance"
	"ciel/internal/display"
)

func unlockInst() {
	basePath := flagCielDir()
	instName := flagInstance()
	parse()

	i := &ciel.Ciel{BasePath: *basePath}
	i.Check()
	c := i.Container()

	if *instName == "" {
		for _, inst := range c.GetAll() {
			_unlockInstEx(inst)
		}
		return
	}

	c.CheckInst(*instName)

	_unlockInstEx(c.Instance(*instName))
}

func shutdownCiel() {
	basePath := flagCielDir()
	parse()

	i := &ciel.Ciel{BasePath: *basePath}
	i.Check()
	c := i.Container()

	for _, inst := range c.GetAll() {
		_shutdownInst(inst)
	}
}

func _shutdownInst(i *instance.Instance) {
	d.SECTION("Instance: " + i.Name)
	d.ITEM("stop")
	tryStop(i)
	d.ITEM("unmount")
	tryUnmount(i)
}

func _unlockInstEx(i *instance.Instance) {
	d.SECTION("Instance: " + i.Name)
	d.ITEM("stop")
	tryStop(i)
	d.ITEM("remove boot flag")
	tryRemove(i.BootLock())
	d.ITEM("remove machine id")
	tryRemove(i.MachineIdFile())
	d.ITEM("remove boot lock")
	tryRemove(i.RefractoryLock())
	d.ITEM("unmount")
	tryUnmount(i)
	d.ITEM("remove lock")
	tryRemove(i.FileSystemLock())
	d.ITEM("remove mount point")
	tryRemove(i.MountPoint())
}

func tryStop(i *instance.Instance) {
	err := i.Stop(context.TODO())
	if err == nil {
		d.OK()
		return
	}
	if err == instance.ErrNoMachineId {
		d.SKIPPED()
		return
	}
	d.FAILED_BECAUSE(err.Error())
}

func tryRemove(path string) {
	err := os.Remove(path)
	if err == nil {
		d.OK()
		return
	}
	if os.IsNotExist(err) {
		d.SKIPPED()
		return
	}
	d.FAILED_BECAUSE(err.Error())
}

func tryUnmount(i *instance.Instance) {
	err := i.Unmount()
	if err == nil {
		d.OK()
		return
	}
	if os.IsNotExist(err) {
		d.SKIPPED()
		return
	}
	d.FAILED_BECAUSE(err.Error())
}
