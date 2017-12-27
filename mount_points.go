package main

import (
	"log"

	"ciel/internal/ciel"
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

func unmountCiel() {
	basePath := flagCielDir()
	instName := flagInstance()
	parse()

	i := &ciel.Ciel{BasePath: *basePath}
	i.Check()
	c := i.Container()

	if *instName == "" {
		instList := c.GetAll()
		for _, inst := range instList {
			err := inst.Unmount()
			if err != nil {
				log.Println(inst.Name+":", err)
			}
		}
		return
	}
	c.CheckInst(*instName)
	c.Instance(*instName).Unmount()
}
