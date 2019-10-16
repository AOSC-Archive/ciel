package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"syscall"

	d "ciel/display"
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
	}

	suffix := " of UNDERLYING OS"
	if !global {
		suffix = ""
	}

	tc := packaging.DetectToolChain(global, inst, c)
	if tc.ACBS {
		packaging.SetTreePath(global, inst, c, pkgtree.TreePath)
	}

	var person string
	if tc.AB {
		if !*batch {
			for person == "" {
				person = d.ASK("Maintainer Info"+suffix, "Foo Bar <myname@example.com>")
			}
		} else {
			person = "Bot <discussions@lists.aosc.io>"
		}
		packaging.SetMaintainer(global, inst, c, person)
	}

	if *batch || d.ASKLower("Would you like to disable DNSSEC feature"+suffix+"?", "yes/no") == "yes" {
		packaging.DisableDNSSEC(global, inst, c)
	}

	if !*batch && d.ASKLower("Would you like to edit sources.list"+suffix+"?", "yes/no") == "yes" {
		packaging.EditSourceList(global, inst, c)
	}

	if !*batch && d.ASKLower("Do you want to enable local packages repository?", "yes/no") == "yes" {
		packaging.InitLocalRepo(global, inst, c)
		// add the key to the APT trust store
		d.ITEM("creating and importing gpg keys")
		if refreshLocalRepo(inst.MountPoint()) == 0 {
			d.OK()
		} else {
			d.FAILED()
		}
	}
}

func refreshLocalRepo(debsDir string) int {
	proc := filepath.Join(PluginDir, PluginPrefix+"localrepo")
	cmd := exec.Command(proc, debsDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if exitErr, ok := err.(*exec.ExitError); ok {
		return exitErr.Sys().(syscall.WaitStatus).ExitStatus()
	}
	if err != nil {
		log.Fatalln(err)
	}
	return 0
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

	debsDir := path.Join(i.GetBasePath(), "OUTPUT", "debs")
	dir, err := os.Getwd()
	debsDirTarget := path.Join(dir, inst.MountPoint(), "debs")
	if err != nil {
		log.Fatalln(err)
	}
	err = os.MkdirAll(debsDir, 0755)
	if err != nil {
		log.Fatalln(err)
	}
	err = os.MkdirAll(debsDirTarget, 0755)
	if err != nil {
		log.Fatalln(err)
	}
	err = syscall.Mount(debsDir, debsDirTarget, "", syscall.MS_BIND, "")
	if err != nil {
		log.Fatalln(err)
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

	err = syscall.Unmount(debsDirTarget, 0)
	if err != nil {
		log.Fatalln(err)
	}
	aptConfigPath := path.Join(inst.MountPoint(), packaging.DefaultRepoConfig)
	if _, err = os.Stat(aptConfigPath); err != nil {
		return
	}

	d.Println(d.C0(d.WHITE, "Refreshing local repository... "))
	refreshLocalRepo(debsDir)
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
