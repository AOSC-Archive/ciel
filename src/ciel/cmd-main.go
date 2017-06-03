package main

import (
	"ci"
	"log"
	"os"
	"os/exec"
	"strings"
)

func cielshell(container *ci.ContainerInstance, args []string) error {
	cmd := container.Exec(ShellPath)
	redirectStdIO(cmd)
	return cmd.Run()
}
func cielrun(container *ci.ContainerInstance, args []string) error {
	cmdline := strings.Join(args, " ")
	arg := []string{ShellPath, "--login", "-c", cmdline}
	cmd := container.Exec(arg...)
	redirectStdIO(cmd)
	return cmd.Run()
}
func cielbuild(container *ci.ContainerInstance, args []string) error {
	if err := cielrbuild(container, args); err != nil {
		return err
	}
	if err := cielcollect(container, []string{}); err != nil {
		return err
	}
	if err := cielclean(container, []string{}); err != nil {
		return err
	}
	return nil
}
func cieltbuild(container *ci.ContainerInstance, args []string) error {
	if err := cielrbuild(container, args); err != nil {
		return err
	}
	if err := cielshell(container, []string{}); err != nil {
		return err
	}
	return nil
}
func cielrbuild(container *ci.ContainerInstance, args []string) error {
	// TODO: handling multi-package building
	arg := []string{ACBSPath, "--clear"}
	arg = append(arg, args[0])
	if err := cielrun(container, arg); err != nil {
		return err
	}
	return nil
}
func cielcollect(container *ci.ContainerInstance, args []string) error {
	const CollectedDir = "collected"
	const ReportDir = CollectedDir + "/report"
	const ReportFile = CollectedDir + "/report.tar.xz"
	const PackageDir = CollectedDir + "/pkg"
	var Targets = []string{
		"amd64",
		"noarch", // FIXME: other possibilities
	}
	fs := container.FS
	os.RemoveAll(CollectedDir)
	os.Mkdir(CollectedDir, 0755)
	os.Mkdir(ReportDir, 0755)
	collectlist := [][2]string{
		{"/var/log/apt/history.log", "/apt-history.log"},
		{"/var/log/apt/term.log", "/apt-term.log"},
		{"/var/log/acbs/acbs-build.log", "/acbs-build.log"},
	}
	for _, pair := range collectlist {
		if err := os.Rename(fs.UpperDir(pair[0]), ReportDir+pair[1]); err == nil {
			log.Println("collect:", pair[0], "->", pair[1])
		}
	}
	if err := os.Rename(fs.UpperDir(findBuildLog(ReportDir+"/acbs-build.log")), ReportDir+"/autobuild.log"); err == nil {
		log.Println("collect:", findBuildLog(ReportDir+"/acbs-build.log"), "->", "/autobuild.log")
	}
	os.Mkdir(CollectedDir+"/pkg", 0755)
	for _, target := range Targets {
		if err := os.Rename(fs.UpperDir("/os-"+target), PackageDir+"/os-"+target); err == nil {
			log.Println("collect: move", "/os-"+target)
		}
	}
	tarcmd := exec.Command("/bin/tar", "-caf", ReportFile, ReportDir)
	if err := tarcmd.Run(); err == nil {
		log.Println("collect: packed report to", ReportFile)
		os.RemoveAll(ReportDir)
	}
	return nil
}
func cielclean(container *ci.ContainerInstance, args []string) error {
	fs := container.FS
	targets := [][]string{
		{"/var/cache/apt", fs.Cache},
		{"/var/cache/acbs/tarballs", fs.Cache},
		{"/var/lib/apt", fs.Cache},
		{"/etc/apt", fs.StubConfig},
	}
	for _, target := range targets {
		fs.Merge(target[0], target[1])
	}
	os.RemoveAll(fs.UpperDir("/"))
	return nil
}
