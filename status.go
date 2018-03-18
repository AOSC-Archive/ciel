package main

import (
	"fmt"

	"ciel/display"
	"ciel/internal/ciel"
)

func list() {
	basePath := flagCielDir()
	parse()

	i := &ciel.Ciel{BasePath: *basePath}
	i.Check()
	c := i.Container()

	fmt.Printf("%s\t%s\t%s\t%s\n", "INSTANCE", "WORKDIR", "STATUS", "BOOT")
	for _, inst := range c.GetAll() {
		showInst := inst.Name
		tabs := "\t\t"
		if len(showInst) > 7 {
			tabs = "\t"
		}
		if len(showInst) > 14 {
			showInst = showInst[:12] + ".."
		}
		var fsStatus, ctnStatus, boot string
		if inst.Running() {
			ctnStatus = d.C0(d.GREEN, "running")
			if inst.RunningAsBootMode() {
				boot = d.C(d.CYAN, "yes")
			} else {
				boot = d.C(d.PURPLE, "no")
			}
		} else {
			ctnStatus = d.C0(d.WHITE, "offline")
		}
		if inst.Mounted() {
			fsStatus = d.C0(d.GREEN, "mounted")
		} else {
			fsStatus = "free"
		}
		fmt.Printf("%s%s%s\t%s\t%s\n", showInst, tabs, fsStatus, ctnStatus, boot)
	}
	fmt.Println()
	count := len(c.GetAllNames())
	if count <= 1 {
		fmt.Printf("%d instance listed.\n", count)
	} else {
		fmt.Printf("%d instances listed.\n", count)
	}
}
