package main

import (
	"log"

	"ciel/internal/ciel"
	"ciel/internal/display"
)

func mountCiel() {
	basePath := flagCielDir()
	instName := flagInstance()
	parse()

	i := &ciel.Ciel{BasePath: *basePath}
	i.Check()
	c := i.Container()

	if *instName == "" {
		instList := c.GetAll()
		for _, inst := range instList {
			err := inst.Mount()
			if err != nil {
				log.Println(inst.Name+":", err)
			}
		}
		return
	}
	c.CheckInst(*instName)
	c.Instance(*instName).Mount()
}

func shutdown() {
	basePath := flagCielDir()
	instName := flagInstance()
	parse()

	i := &ciel.Ciel{BasePath: *basePath}
	i.Check()
	c := i.Container()

	if *instName == "" {
		instList := c.GetAll()
		for _, inst := range instList {
			d.SECTION("Shutdown Instance " + inst.Name)
			inst.Unmount()
		}
		return
	}
	c.CheckInst(*instName)
	d.SECTION("Shutdown Instance " + *instName)
	c.Instance(*instName).Unmount()
}
