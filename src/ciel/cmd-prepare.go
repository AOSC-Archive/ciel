package main

import (
	"bufio"
	"ci"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
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
	dict, err := getPkgFiles(container)
	if err != nil {
		return err
	}
	fs := container.FS
	err = filepath.Walk(fs.UpperDir("/etc"), func(path string, info os.FileInfo, err error) error {
		if info == nil {
			return err
		}
		rel, err := filepath.Rel(fs.UpperDir("/"), path)
		if err != nil {
			return err
		}
		rel = "/" + rel
		_, ok := dict[rel]
		if !ok {
			log.Println(rel)
		}
		return nil
	})
	return err
}
func getPkgFiles(container *ci.ContainerInstance) (map[string]bool, error) {
	cmdline := `dpkg-query --listfiles $(dpkg-query --show --showformat=\$"{Package}\n")`
	arg := []string{ShellPath, "--login", "-c", cmdline}
	cmd := container.Exec(arg...)
	stdout, err := cmd.StdoutPipe()
	stdoutbuf := bufio.NewReaderSize(stdout, 1*1024*1024)
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	dict := make(map[string]bool, 50000)
	var line string
	var loop int
	for {
		slice, isPrefix, err := stdoutbuf.ReadLine()
		if err != nil {
			break
		}
		line = line + string(slice)
		if !isPrefix {
			dict[line] = true
			if loop%200 == 0 {
				fmt.Printf("unique path: %d\r", len(dict))
			}
			line = ""
		}
	}
	fmt.Printf("unique path: %d\n", len(dict))
	if err := cmd.Wait(); err != nil {
		return nil, err
	}
	return dict, nil
}
