package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"ciel/internal/cieldir.1"
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
		instName := requireEnv("CIEL_INST")
		if i.InstExists(instName) {
			log.Fatalln("already has " + instName)
		}
		i.InstAdd(instName)
		i.InstMount(instName)
	case "del":
		instName := requireEnv("CIEL_INST")
		if !i.InstExists(instName) {
			log.Fatalln("could not find " + instName)
		}
		i.InstUnmount(instName)
		i.InstDel(instName)
	case "start":
		i.MountAll()
	case "stop":
		i.UnmountAll()
	case "unlock-filesystem":
		instName := requireEnv("CIEL_INST")
		fmt.Println("Warning: you should only use this when a unmounted insntance has not been unlocked")
		fmt.Print("Continue? y/n: ")
		var answer string
		fmt.Scanln(&answer)
		if answer == "y" {
			log.Println(i.InstUnmount(instName))
			log.Println(os.Remove(i.InstLockFile(instName)))
			log.Println(os.Remove(i.InstMountPoint(instName)))
		} else {
			fmt.Println("Cancelled.")
		}
	case "down":
		instName := requireEnv("CIEL_INST")
		err := i.InstStop(context.TODO(), instName)
		if err != nil {
			log.Fatal(err)
		}
	case "unlock-container":
		instName := requireEnv("CIEL_INST")
		fmt.Println("Warning: you should only use this when a stopped insntance has not been unlocked")
		fmt.Print("Continue? y/n: ")
		var answer string
		fmt.Scanln(&answer)
		if answer == "y" {
			//log.Println(i.InstPoweroff(instName))
			log.Println(os.Remove(i.InstBootedFile(instName)))
			log.Println(os.Remove(i.InstMachineIdFile(instName)))
			log.Println(os.Remove(i.InstRefractoryFile(instName)))
		} else {
			fmt.Println("Cancelled.")
		}
	case "", "list":
		for _, inst := range i.GetAll() {
			fmt.Println(inst + "\t" + i.InstLockStat(inst))
		}
	case "shell":
		instName := requireEnv("CIEL_INST")
		exitStatus, err := i.InstRun(context.TODO(), instName, true, nil, args[1], args[2:]...)
		if err != nil {
			log.Println(err)
			os.Exit(exitStatus)
		}
	default:
		log.Fatalln("unknown command")
	}
}

func requireArg(i int) string { // TODO: package flag
	if i >= len(args) {
		log.Fatalln("not enough arguments")
	}
	return args[i]
}

func requireEnv(env string) string {
	v, ok := os.LookupEnv(env)
	if !ok {
		log.Fatalln("not enough arguments")
	}
	return v
}
