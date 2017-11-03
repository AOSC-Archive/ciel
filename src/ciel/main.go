package main

import (
	"flag"
	"log"
	"fmt"

	"ciel/src/ciel/internal/cieldir.1"
)

var args []string

func main() {
	flag.Parse()
	args = flag.Args()
	parser()
}

func parser() {
	cmd := ""
	if len(args) > 0 {
		cmd = args[0]
	}
	i := &cieldir.CielDir{}
	i.Check()
	switch cmd {
	case "init":
		i.Init()
	case "load":
		Untar(i, requireArg(1))
	case "add":
		instName := requireArg(1)
		if i.HasInst(instName) {
			log.Fatalln("already has " + instName)
		}
		i.CreateInst(instName)
		i.MountInst(instName)
	case "del":
		instName := requireArg(1)
		if !i.HasInst(instName) {
			log.Fatalln("could not find " + instName)
		}
		i.UnmountInst(instName)
		i.DeleteInst(instName)
	case "", "start":
		i.MountAll()
	case "stop":
		i.UnmountAll()
	case "list":
		for _, inst := range i.AllInst() {
			fmt.Println(inst + "\t" + i.InstStat(inst))
		}
	default:
		log.Fatalln("unknown command")
	}
}

func requireArg(i int) string {
	if i >= len(args) {
		log.Fatalln("not enough arguments")
	}
	return args[i]
}