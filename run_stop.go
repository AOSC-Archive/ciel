package main

import (
	"context"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	"ciel/internal/cieldir.1"
)

func shell() {
	rawRun(true)
}

func run() {
	rawRun(false)
}

func rawRun(shell bool) {
	basePath := flagCielDir()
	instName := flagInstance()
	networkFlag := flagNetwork()
	noBooting := flagNoBooting()
	bootConfig := flagBootConfig()
	parse()

	i := &cieldir.CielDir{BasePath: *basePath}
	i.Check()
	i.CheckInst(*instName)

	i.InstMount(*instName)

	var args []string

	bootConf := strings.Split(strings.TrimSpace(*bootConfig), "\n")

	if shell {
		rootShell := "/bin/sh"
		passwdFileName := path.Join(i.InstMountPoint(*instName), "/etc/passwd")
		a, err := ioutil.ReadFile(passwdFileName)
		if err != nil {
			log.Panic(err)
		}
		passwd := string(a)
		for _, userInfo := range strings.Split(passwd, "\n") {
			if userInfo == "" {
				continue
			}
			fields := strings.Split(userInfo, ":")
			if fields[0] == "root" {
				rootShell = fields[6]
			}
		}
		if flag.NArg() > 1 {
			log.Fatalln("you must pass one argument only")
		}
		if cmd := flag.Arg(0); cmd != "" {
			args = []string{
				rootShell,
				"--login",
				"-c", cmd,
			}
		} else {
			args = []string{
				rootShell,
			}
		}
	} else {
		args = flag.Args()
	}

	exitStatus, err := i.InstRun(
		context.TODO(),
		*instName,
		!*noBooting,
		*networkFlag,
		bootConf,
		args...,
	)

	if err != nil {
		log.Println(err)
	}
	os.Exit(exitStatus)
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
