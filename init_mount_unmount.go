package main

import "ciel/internal/container/dotciel.1"

func initCiel() {
	basePath := flagCielDir()
	parse()

	i := &dotciel.CielDir{BasePath: *basePath}
	i.Init()
}

func mountCiel() {
	basePath := flagCielDir()
	instName := flagInstance()
	parse()

	i := &dotciel.CielDir{BasePath: *basePath}
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

	i := &dotciel.CielDir{BasePath: *basePath}
	i.Check()

	if *instName == "" {
		i.UnmountAll()
		return
	}
	i.CheckInst(*instName)
	i.InstUnmount(*instName)
}
