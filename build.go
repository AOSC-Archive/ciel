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
	"ciel/internal/packaging"
	"ciel/internal/pkgtree"
)

func buildConfig() {
	basePath := flagCielDir()
	instName := flagInstance()
	batch := flagBatch()
	parse()

	i := &ciel.Ciel{BasePath: *basePath}
	i.Check()
	c := i.Container()
	c.CheckInst(*instName)

	inst := c.Instance(*instName)
	inst.Unmount()
	inst.MountLocal()
	defer func() {
		inst.Unmount()
	}()

	packaging.DetectToolChain(inst)
	packaging.SetTreePath(inst, pkgtree.TreePath)
	var person string
	if !*batch {
		for person == "" {
			person = d.ASK("Maintainer Info", "Foo Bar <myname@example.com>")
		}
	} else {
		person = "Bot <discussions@lists.aosc.io>"
	}
	packaging.SetMaintainer(inst, person)
	if !*batch && d.ASK("Would you like to disable DNSSEC feature?", "yes/no") == "yes" {
		packaging.DisableDNSSEC(inst)
	}
	if !*batch && d.ASK("Would you like to edit source list?", "yes/no") == "yes" {
		packaging.EditSourceList(inst)
	}
}

func build() {
	basePath := flagCielDir()
	instName := flagInstance()
	networkFlag := flagNetwork()
	noBooting := flagNoBooting()
	containerArg := flagArgs()
	parse()

	i := &ciel.Ciel{BasePath: *basePath}
	i.Check()
	c := i.Container()
	c.CheckInst(*instName)
	inst := c.Instance(*instName)
	inst.Mount()

	containerArgs := strings.Split(strings.TrimSpace(*containerArg), "\n")

	shell, err := inst.Shell("root")
	if err != nil {
		log.Fatal(err)
	}
	args := []string{
		shell,
		"--login",
		"-c", `acbs-build "` + flag.Arg(0) + `"`,
	}

	exitStatus, err := inst.Run(context.Background(),
		!*noBooting,
		*networkFlag,
		containerArgs,
		args...,
	)
	if err != nil {
		log.Fatalln(err)
	}
	if exitStatus != 0 {
		os.Exit(exitStatus)
	}

	os.Mkdir(path.Join(i.GetBasePath(), "OUTPUT"), 0755)
	cmd := exec.Command("sh", "-c", "cp -rp "+inst.MountPoint()+"/os-* OUTPUT/")
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
