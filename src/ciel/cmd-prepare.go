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
func cielinitAptFullUpgrade(container *ci.ContainerInstance, tarball string) error {
	log.Println("init.2: apt full-upgrade")
	container.NoBooting = true
	container.Cmd = "/usr/bin/apt"
	container.Args = []string{
		"full-upgrade",
		"--yes",
	}
	return nil
}
func cielinitAptInstallSystemd(container *ci.ContainerInstance, tarball string) error {
	log.Println("init.3: apt install systemd")
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
	log.Println("init.4: apt install bases")
	aptInstall := []string{"apt", "install", "--yes"}
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
	log.Println("init.4: systemctl disable units")
	disableUnits := []string{
		"fc-cache",
	}
	for _, unit := range disableUnits {
		if err := cielrun(container, append([]string{"systemctl", "disable"}, unit)); err != nil {
			log.Println("init: disable unit "+unit+" failed", err)
		}
	}
	log.Println("init.4: FIXME: clean up")
	return nil
}
func cielupdate(container *ci.ContainerInstance, args []string) error {
	return errors.New("not implemented")
}
