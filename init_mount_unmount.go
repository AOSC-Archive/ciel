package main

import "ciel/internal/cieldir.1"

func initCiel() {
	basePath := flagCielDir()
	parse()

	i := &cieldir.CielDir{BasePath: *basePath}
	i.Check()
	i.Init()
}

func mountCiel() {
	basePath := flagCielDir()
	instName := flagInstance()
	parse()

	i := &cieldir.CielDir{BasePath: *basePath}
	i.Check()

	if *instName == "" {
		i.MountAll()
		return
	}
	i.CheckInst(*instName)
	i.InstMount(*instName)
}

func unmountCiel() {
	basePath := flagCielDir()
	instName := flagInstance()
	parse()

	i := &cieldir.CielDir{BasePath: *basePath}
	i.Check()

	if *instName == "" {
		i.UnmountAll()
		return
	}
	i.CheckInst(*instName)
	i.InstUnmount(*instName)
}
