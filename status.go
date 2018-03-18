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

	fmt.Printf("\t%s\t%s\t%s\t%s\n", "INSTANCE", "WORKDIR", "STATUS", "BOOT")
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
			ctnStatus = d.C0(d.CYAN, "running")
		} else {
			ctnStatus = d.C0(d.GREEN, "offline")
		}
		if inst.RunningAsBootMode() {
			boot = d.C0(d.CYAN, "yes")
		} else {
			boot = d.C0(d.WHITE, "no")
		}
		if inst.Mounted() {
			fsStatus = d.C0(d.GREEN, "mounted")
		} else {
			fsStatus = "free"
		}
		fmt.Printf("\t%s%s%s\t%s\t%s\n", showInst, tabs, fsStatus, ctnStatus, boot)
	}
	fmt.Println()
	count := len(c.GetAllNames())
	if count <= 1 {
		fmt.Printf("\t%d instance listed.\n", count)
	} else {
		fmt.Printf("\t%d instances listed.\n", count)
	}
}
