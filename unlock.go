package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"ciel/internal/cieldir.1"
)

func unlockInst() {
	basePath := flagCielDir()
	instName := flagInstance()
	parse()

	i := &cieldir.CielDir{BasePath: *basePath}
	i.Check()
	i.CheckInst(*instName)

	fmt.Println("Warning: you should only use this when a stopped insntance has not been unlocked")
	fmt.Print("Continue? y/n: ")
	var answer string
	fmt.Scanln(&answer)
	if answer == "y" {
		log.Println(i.InstStop(context.TODO(), *instName))
		log.Println(os.Remove(i.InstBootedFile(*instName)))
		log.Println(os.Remove(i.InstMachineIdFile(*instName)))
		log.Println(os.Remove(i.InstRefractoryFile(*instName)))
		log.Println(i.InstUnmount(*instName))
		log.Println(os.Remove(i.InstLockFile(*instName)))
		log.Println(os.Remove(i.InstMountPoint(*instName)))
	} else {
		fmt.Println("Cancelled.")
	}
}
