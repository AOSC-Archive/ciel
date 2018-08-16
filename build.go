package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"

	"ciel/display"
	"ciel/internal/ciel"
	"ciel/internal/container/instance"
	"ciel/internal/packaging"
	"ciel/internal/pkgtree"
)

func buildConfig() {
	basePath := flagCielDir()
	instName := flagInstance()
	batch := flagBatch()
	var global = false
	flag.BoolVar(&global, "g", global, "global, configure for underlying OS")
	parse()

	i := &ciel.Ciel{BasePath: *basePath}
	i.Check()
	c := i.Container()

	var inst *instance.Instance

	if !global {
		c.CheckInst(*instName)
		inst = c.Instance(*instName)
		inst.Unmount()
		inst.MountLocal()
		defer func() {
			inst.Unmount()
		}()

		tc := packaging.DetectToolChain(inst)
		if tc.ACBS {
			packaging.SetTreePath(inst, pkgtree.TreePath)
		}

		var person string
		if tc.AB {
			if !*batch {
				for person == "" {
					person = d.ASK("Maintainer Info", "Foo Bar <myname@example.com>")
				}
			} else {
				person = "Bot <discussions@lists.aosc.io>"
			}
			packaging.SetMaintainer(inst, person)
		}
		if *batch || d.ASKLower("Would you like to disable DNSSEC feature?", "yes/no") == "yes" {
			packaging.DisableDNSSEC(inst)
		}
	}

	suffix := " of UNDERLYING OS"
	if !global {
		suffix = ""
	}
	if !*batch && d.ASKLower("Would you like to edit sources.list"+suffix+"?", "yes/no") == "yes" {
		packaging.EditSourceList(global, inst, c)
	}
}

func build() {
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

	rootShell, err := inst.Shell("root")
	if err != nil {
		log.Fatal(err)
	}
	args := []string{
		rootShell,
		"--login",
		"-c", `acbs-build ` + strings.Join(flag.Args(), " "),
		// FIXME: 'strict mode' -- only one package, require Local DEB Repository
	}

	ctnInfo := buildContainerInfo(!*noBooting, *networkFlag)
	runInfo := buildRunInfo(args)

	exitStatus, err := inst.Run(
		context.Background(),
		ctnInfo,
		runInfo,
	)
	if err != nil {
		log.Fatalln(err)
	}
	if exitStatus != 0 {
		os.Exit(exitStatus)
	}

	os.Mkdir(path.Join(i.GetBasePath(), "OUTPUT"), 0755)
	cmd := exec.Command("sh", "-c", "cp -rp "+inst.MountPoint()+"/debs OUTPUT/")
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if exitErr, ok := err.(*exec.ExitError); ok {
		os.Exit(exitErr.Sys().(syscall.WaitStatus).ExitStatus())
	}
	if err != nil {
		log.Fatalln(err)
	}

	//cmd = exec.Command("sh", "-c", "cp -p "+inst.MountPoint()+"/var/log/apt/history.log OUTPUT/")
	//cmd.Stderr = os.Stderr
	//err = cmd.Run()
	//if exitErr, ok := err.(*exec.ExitError); ok {
	//	os.Exit(exitErr.Sys().(syscall.WaitStatus).ExitStatus())
	//}
	//if err != nil {
	//	log.Fatalln(err)
	//}

	// TODO: collect information
}
