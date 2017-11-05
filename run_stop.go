package main

import (
	"context"
	"flag"
	"log"
	"os"
	"strings"

	"ciel/internal/cieldir.1"
)

func run() {
	basePath := flagCielDir()
	instName := flagInstance()
	noBootFlag := getEnv("CIEL_BOOT", "false") == "true"
	flag.BoolVar(&noBootFlag, "n", noBootFlag, "do not boot the container")
	bootConfString := getEnv("CIEL_BOOTCFG", "")
	parse()

	i := &cieldir.CielDir{BasePath: *basePath}
	i.Check()
	i.CheckInst(*instName)

	bootConf := strings.Split(strings.TrimSpace(bootConfString), "\n")
	exitStatus, err := i.InstRun(context.TODO(), *instName, !noBootFlag, bootConf, flag.Args()...)
	if err != nil {
		log.Println(err)
		os.Exit(exitStatus)
	}
}

func stop() {
	basePath := flagCielDir()
	instName := flagInstance()
	parse()

	i := &cieldir.CielDir{BasePath: *basePath}
	i.Check()
	i.CheckInst(*instName)

	err := i.InstStop(context.TODO(), *instName)
	if err != nil {
		log.Fatal(err)
	}
}
