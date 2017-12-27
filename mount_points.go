package main

import (
	"log"

	"ciel/internal/dotciel"
)

func mountCiel() {
	basePath := flagCielDir()
	instName := flagInstance()
	parse()

	i := &dotciel.Ciel{BasePath: *basePath}
	i.Check()
	c := i.Container()

	if *instName == "" {
		instList := c.GetAll()
		for _, inst := range instList {
			err := inst.InstMount()
			if err != nil {
				log.Println(inst.Name+":", err)
			}
		}
		return
	}
	c.CheckInst(*instName)
	c.Instance(*instName).InstMount()
}

func unmountCiel() {
	basePath := flagCielDir()
	instName := flagInstance()
	parse()

	i := &dotciel.Ciel{BasePath: *basePath}
	i.Check()
	c := i.Container()

	if *instName == "" {
		instList := c.GetAll()
		for _, inst := range instList {
			err := inst.InstUnmount()
			if err != nil {
				log.Println(inst.Name+":", err)
			}
		}
		return
	}
	c.CheckInst(*instName)
	c.Instance(*instName).InstUnmount()
}
