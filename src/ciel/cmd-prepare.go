package main

import (
	"ci"
	"errors"
	"log"
	"os/exec"
)

func cielinitAptUpdate(container *ci.ContainerInstance, tarball string) error {
	log.Println("init.1: extract stub into", container.FS.Stub)
	tarcmd := exec.Command("/bin/tar", "-xf", tarball, "-C", container.FS.Stub)
	if err := tarcmd.Run(); err != nil {
		return errors.New("init.1: " + err.Error())
	}
	log.Println("init.1: apt update")
	container.NoBooting = true
	container.Cmd = "/usr/bin/apt"
	container.Args = []string{
		"update",
	}
	return nil
}
func cielinitAptInstallSystemd(container *ci.ContainerInstance, tarball string) error {
	log.Println("init.2: apt install systemd")
	container.NoBooting = true
	container.Cmd = "/usr/bin/apt"
	container.Args = []string{
		"install",
		"--yes",
		"systemd",
	}
	return nil
}
func cielpostinit(container *ci.ContainerInstance, args []string) error {
	if container.NoBooting {
		return nil // end of Stage 1
	}
	aptInstall := []string{"apt", "install"}
	pkgs := []string{
		"admin-base",
		"core-base",
		"editor-base",
		"python-base",
		"network-base",
		"systemd-base",
		"web-base",
		"util-base",
		"devel-base",
		"debug-base",
		"autobuild3",
		// "acbs",
		// TODO: aosc-os-abbs: add acbs.
		"git",
	}
	if err := cielrun(container, append(aptInstall, pkgs...)); err != nil {
		return err
	}
	disableUnits := []string{}
	for _, unit := range disableUnits {
		if err := cielrun(container, append([]string{"systemctl", "disable"}, unit)); err != nil {
			log.Println("init: disable unit "+unit+" failed", err)
		}
	}
	log.Println("init: FIXME: clean up")
	return nil
}
func cielupdate(container *ci.ContainerInstance, args []string) error {
	return errors.New("not implemented")
}
