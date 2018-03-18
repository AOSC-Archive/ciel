package main

import (
	"context"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
	"strings"

	"ciel/display"
	"ciel/internal/ciel"
	"ciel/internal/container/instance"
)

func add() {
	basePath := flagCielDir()
	parse()
	instName := flag.Arg(0)

	if instName == "" {
		log.Fatalln("give me a name for the new instance")
	}
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
	basePath := flagCielDir()
	instName := flagInstance()
	networkFlag := flagNetwork()
	noBooting := flagNoBooting()
	parse()

	if flag.NArg() > 1 {
		log.Fatalln("you must pass one argument only")
	}

	i := &ciel.Ciel{BasePath: *basePath}
	i.Check()
	c := i.Container()
	c.CheckInst(*instName)

	inst := c.Instance(*instName)
	inst.Mount()

	if flag.NArg() == 0 {
		exitStatus, err := _openShell(
			inst,
			*networkFlag,
			!*noBooting,
		)
		if err != nil {
			log.Println(err)
		}
		os.Exit(exitStatus)
	}
	exitStatus, err := _shellRun(
		inst,
		*networkFlag,
		!*noBooting,
		flag.Arg(0),
	)
	if err != nil {
		log.Println(err)
	}
	os.Exit(exitStatus)
}

func run() {
	basePath := flagCielDir()
	instName := flagInstance()
	networkFlag := flagNetwork()
	noBooting := flagNoBooting()
	parse()

	i := &ciel.Ciel{BasePath: *basePath}
	i.Check()
	c := i.Container()
	c.CheckInst(*instName)

	inst := c.Instance(*instName)
	inst.Mount()

	ctnInfo := buildContainerInfo(!*noBooting, *networkFlag)
	runInfo := buildRunInfo(flag.Args())
	runInfo.UseSystemdRun = true
	exitStatus, err := inst.Run(
		context.TODO(),
		ctnInfo,
		runInfo,
	)

	if err != nil {
		log.Println(err)
	}
	os.Exit(exitStatus)
}

func _openShell(inst *instance.Instance, network bool, boot bool) (int, error) {
	inst.Mount()
	rootShell, err := inst.Shell("root")
	if err != nil {
		return -1, err
	}
	ctnInfo := buildContainerInfo(boot, network)
	runInfo := buildRunInfo([]string{rootShell})

	exitStatus, err := inst.Run(
		context.TODO(),
		ctnInfo,
		runInfo,
	)
	if err != nil {
		return -1, err
	}
	return exitStatus, nil
}

func _shellRun(inst *instance.Instance, network bool, boot bool, cmd string) (int, error) {
	inst.Mount()
	var args []string
	rootShell, err := inst.Shell("root")
	if err != nil {
		return -1, err
	}
	if cmd != "" {
		cmd = "stty -onlcr; " + cmd + "; echo $?>/.ciel-exit-status"
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
	ctnInfo := buildContainerInfo(boot, network)
	runInfo := buildRunInfo(args)
	exitStatus, err := inst.Run(
		context.TODO(),
		ctnInfo,
		runInfo,
	)
	exitStatusFile := path.Join(inst.MountPoint(), ".ciel-exit-status")
	if b, err := ioutil.ReadFile(exitStatusFile); err == nil {
		os.Remove(exitStatusFile)
		if realExitStatus, err := strconv.Atoi(strings.TrimSpace(string(b))); err == nil {
			return realExitStatus, nil
		} else {
			log.Println(err)
			return exitStatus, nil
		}
	} else {
		if os.IsNotExist(err) {
			log.Println("session was accidentally terminated")
		} else {
			log.Println(err)
		}
		if exitStatus == 0 {
			exitStatus = 1
		}
		return exitStatus, nil
	}
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
		inst.Unmount()
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
		inst.Unmount()
	} else {
		d.Println(d.C(d.CYAN, "OFFLINE"))
	}
	inst.FileSystem().Merge()
}
