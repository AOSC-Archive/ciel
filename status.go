package main

import (
	"fmt"

	"ciel/internal/container/dotciel.1"
)

func list() {
	basePath := flagCielDir()
	parse()

	i := &dotciel.CielDir{BasePath: *basePath}
	i.Check()

	fmt.Printf("%s\t%s\t%s\t%s\n", "INSTANCE", "WORKDIR", "STATUS", "BOOT")
	for _, inst := range i.GetAll() {
		showInst := inst
		tabs := "\t\t"
		if len(inst) > 7 {
			tabs = "\t"
		}
		if len(inst) > 14 {
			showInst = inst[:12] + ".."
		}
		status, boot := i.InstContainerRunningStat(inst)
		fmt.Printf("%s%s%s\t%s\t%s\n", showInst, tabs, i.InstFileSystemStat(inst), status, boot)
	}
	fmt.Println()
	count := len(i.GetAll())
	if count <= 1 {
		fmt.Printf("%d instance listed.\n", len(i.GetAll()))
	} else {
		fmt.Printf("%d instances listed.\n", len(i.GetAll()))
	}
}
