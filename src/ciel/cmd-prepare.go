package main

import (
	"ci"
	"errors"
	"log"
	"os/exec"
)

func cielinit(fs *ci.ContainerFilesystem, tarball string) error {
	tarcmd := exec.Command("/bin/tar", "-xf", tarball, fs.Stub)
	if err := tarcmd.Run(); err == nil {
		return errors.New("init: " + err.Error())
	}
	log.Println("init: extracted stub into", fs.Stub)
	return nil
}
func cielpostinit(container *ci.ContainerInstance, args []string) error {
	aptinstall := []string{"apt", "install"}
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
	if err := cielrun(container, append(aptinstall, pkgs...)); err != nil {
		return err
	}
	log.Println("init: FIXME: clean up")
	return nil
}
func cielupdate(container *ci.ContainerInstance, args []string) error {
	return errors.New("not implemented")
}
