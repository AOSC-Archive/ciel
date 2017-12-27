package main

import (
	"context"
	"flag"
	"log"
	"os"
	"strings"

	"ciel/internal/ciel"
	"ciel/internal/display"
)

func add() {
	basePath := flagCielDir()
	parse()
	instName := flag.Arg(0)

	if strings.Contains(instName, " ") {
		log.Fatalln("do not contain white space")
	}

	i := &ciel.Ciel{BasePath: *basePath}
	i.Check()
	c := i.Container()

	if c.InstExists(instName) {
		log.Fatalln("already has " + instName)
	}
	c.AddInst(instName)
	c.Instance(instName).Mount()
}

func del() {
	basePath := flagCielDir()
	parse()
	instName := flag.Arg(0)

	i := &ciel.Ciel{BasePath: *basePath}
	i.Check()
	c := i.Container()
	c.CheckInst(instName)

	c.Instance(instName).Unmount()
	c.DelInst(instName)
}

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

	i := &ciel.Ciel{BasePath: *basePath}
	i.Check()
	c := i.Container()
	c.CheckInst(*instName)

	inst := c.Instance(*instName)
	inst.Mount()

	var args []string

	bootConf := strings.Split(strings.TrimSpace(*bootConfig), "\n")

	if shell {
		rootShell, err := inst.Shell("root")
		if err != nil {
			log.Fatalln(err)
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

	exitStatus, err := inst.Run(
		context.TODO(),
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

	i := &ciel.Ciel{BasePath: *basePath}
	i.Check()
	c := i.Container()
	c.CheckInst(*instName)

	inst := c.Instance(*instName)

	err := inst.Stop(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
}

func rollback() {
	basePath := flagCielDir()
	instName := flagInstance()
	parse()

	i := &ciel.Ciel{BasePath: *basePath}
	i.Check()
	c := i.Container()
	c.CheckInst(*instName)

	inst := c.Instance(*instName)

	d.SECTION("Rollback Changes")
	d.ITEM("is running?")
	if inst.Running() {
		d.Println(d.C(d.YELLOW, "ONLINE"))
	} else {
		d.Println(d.C(d.CYAN, "OFFLINE"))
	}
	inst.FileSystem().Rollback()
}

func commit() {
	basePath := flagCielDir()
	instName := flagInstance()
	parse()

	i := &ciel.Ciel{BasePath: *basePath}
	i.Check()
	c := i.Container()
	c.CheckInst(*instName)

	inst := c.Instance(*instName)

	d.SECTION("Commit Changes")
	d.ITEM("is running?")
	if inst.Running() {
		d.Println(d.C(d.YELLOW, "ONLINE"))
	} else {
		d.Println(d.C(d.CYAN, "OFFLINE"))
	}
	inst.FileSystem().Merge()
}
