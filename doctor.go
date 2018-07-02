package main

import (
	"context"
	"strconv"

	"ciel/display"
	"ciel/internal/ciel"
	"ciel/internal/container/instance"
	"ciel/systemd-api/nspawn"
)

func doctor() {
	basePath := flagCielDir()
	instName := flagInstance()
	parse()

	i := &ciel.Ciel{BasePath: *basePath}
	i.Check()
	c := i.Container()

	var instList []*instance.Instance
	if *instName == "" {
		instList = c.GetAll()
	} else {
		c.CheckInst(*instName)
		instList = []*instance.Instance{c.Instance(*instName)}
	}
	for _, inst := range instList {
		d.SECTION("Status of " + inst.Name)

		d.ITEM("SEMAPHORES")
		fl, flErr := inst.FileSystemLock().Get()
		rl, rlErr := inst.RunLock().Get()
		flName := "file@" + strconv.FormatUint(uint64(inst.FileSystemLock().S), 16)
		rlName := "run@" + strconv.FormatUint(uint64(inst.RunLock().S), 16)
		if flErr == nil && rlErr == nil && fl <= 1 && rl <= 1 {
			d.Print(d.C0(switchColor(fl == 0), flName))
			d.Print(" ")
			d.Print(d.C0(switchColor(rl == 0), rlName))
			d.Println()
		} else {
			d.Println()
			d.ITEM(flName)
			if flErr != nil {
				d.Println(d.C(d.RED, flErr.Error()))
			} else {
				d.Println(d.C(d.RED, strconv.Itoa(fl)))
			}
			d.ITEM(rlName)
			if rlErr != nil {
				d.Println(d.C(d.RED, rlErr.Error()))
			} else {
				d.Println(d.C(d.RED, strconv.Itoa(rl)))
			}
		}

		d.ITEM("OVERLAYFS")
		d.Println(d.C0(switchColor(inst.Mounted()), "mounted"))

		d.ITEM("CONTAINER")
		d.Print(d.C0(switchColor(inst.Running()), "running"))
		d.Print(" ")
		d.Print(d.C0(switchColor(inst.RunningAsBootMode()), "bootMode"))
		d.Print(" ")
		d.Print(d.C0(switchColor(inst.RunningAsExclusiveMode()), "simpleMode"))
		d.Println()

		d.ITEM("SYSTEMD")
		st := nspawn.MachineStatus(context.Background(), inst.MachineId())
		d.Print(d.C0(switchColor(nspawn.MachineDead(st)), "dead"))
		d.Print(" ")
		d.Print(d.C0(switchColor(nspawn.MachineRunning(st)), "running"))
		d.Println()
		d.ITEM("(raw str)")
		lim := len(st) - 1 - 20
		pf := "..."
		if lim < 0 {
			lim = 0
			pf = ""
		}
		d.Println(d.C(d.WHITE, pf+st[lim:]))
	}
}

//func checkUname() {
//	uname := syscall.Utsname{}
//	err := syscall.Uname(&uname)
//}

func switchColor(b bool) d.Color {
	if b {
		return d.CYAN
	}
	return d.WHITE
}
