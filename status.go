package main

import (
	"fmt"

	"ciel/internal/dotciel"
)

func list() {
	basePath := flagCielDir()
	parse()

	i := &dotciel.Ciel{BasePath: *basePath}
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
		status, boot := inst.ContainerStat()
		fmt.Printf("\t%s%s%s\t%s\t%s\n", showInst, tabs, inst.FileSystemStat(), status, boot)
	}
	fmt.Println()
	count := len(c.GetAllNames())
	if count <= 1 {
		fmt.Printf("\t%d instance listed.\n", count)
	} else {
		fmt.Printf("\t%d instances listed.\n", count)
	}
}
