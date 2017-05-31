package main

import (
	"ci"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

const MachineName = "buildkit"
const ShellPath = "/bin/bash"
const ACBSPath = "/bin/acbs-build"

type HandlerFunc func(*ci.ContainerInstance, []string) error
type Route struct {
	name    string
	argc    int // -1 = infinite
	needfs  bool
	needci  bool
	handler HandlerFunc
}

var router = []Route{
	{"shell", 0, true, true, cielshell},
	{"run", -1, true, true, cielrun},
	{"build", 1, true, true, cielbuild},
	{"tbuild", 1, true, true, cieltbuild},
	{"rbuild", 1, true, true, cielrbuild},
	{"collect", 0, false, false, cielcollect},
	{"clean", 0, false, false, cielclean},
}

func main() {
	var route *Route

	if len(os.Args) == 1 {
		cielhelp()
		return
	}
	for _, xroute := range router {
		if os.Args[1] == xroute.name && (xroute.argc == -1 || len(os.Args)-2 == xroute.argc) {
			route = &xroute
			break
		}
	}
	if route == nil {
		cielhelp()
		return
	}

	fs := ci.InitFilesystem("buildkit") // FIXME: configurability
	container := ci.NewContainer(fs, MachineName)

	if route.needfs {
		if err := fs.Startup(); err != nil {
			log.Panicln("fs", err)
		}
		defer func() {
			if err := fs.Shutdown(); err != nil {
				log.Panicln("shutdownfs", err)
			}
		}()

		if route.needci {
			if err := container.Startup(); err != nil {
				log.Panicln("container", err)
			}
			defer func() {
				if err := container.Shutdown(); err != nil {
					log.Panicln("shutdown", err)
				}
			}()
		}
	}

	args := []string{}
	if len(os.Args) > 2 {
		args = os.Args[2:]
	}
	if err := route.handler(container, args); err != nil {
		log.Panicln(err)
	}
}

func cielhelp() {
	fmt.Print(`Usage:	` + os.Args[0] + ` <command> [arg...]

Most used commands:
	(default)       Show this information
	shell           Shell in container
	run     <cmd>   Run command in container
	build   <pkg>   Equivalent to "rbuild + collect + clean"
	tbuild  <pkg>   Build package, but stay in shell to test the package

Underlying operations:
	rbuild  <pkg>   Build package
	collect         Collect packaging output and log files
	clean           Merge cache to "overlay" and reset container
`)
}

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
		if err := os.Rename(fs.DiffDir(pair[0]), ReportDir+pair[1]); err == nil {
			log.Println("collect:", pair[0], "->", pair[1])
		}
	}
	if err := os.Rename(fs.DiffDir(findBuildLog(ReportDir+"/acbs-build.log")), ReportDir+"/autobuild.log"); err == nil {
		log.Println("collect:", findBuildLog(ReportDir+"/acbs-build.log"), "->", "/autobuild.log")
	}
	os.Mkdir(CollectedDir+"/pkg", 0755)
	for _, target := range Targets {
		if err := os.Rename(fs.DiffDir("/os-"+target), PackageDir+"/os-"+target); err == nil {
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
	targets := []string{
		"/var/cache/apt",
		"/var/cache/acbs/tarballs",
		"/var/lib/apt",
		"/etc/apt",
	}
	for _, target := range targets {
		fs.Merge(target)
	}
	os.RemoveAll(fs.DiffDir("/"))
	return nil
}
