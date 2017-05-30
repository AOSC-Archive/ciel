package main

import (
	"ci"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

const MachineName = "buildkit"
const ShellPath = "/bin/bash"
const ACBSPath = "/bin/acbs-build"

func main() {
	var cielrouter func(*ci.ContainerInstance, []string) error
	switch {
	case len(os.Args) == 1+1 && os.Args[1] == "shell":
		cielrouter = cielshell
	case len(os.Args) >= 1+2 && os.Args[1] == "run":
		cielrouter = cielrun
	case len(os.Args) >= 1+2 && os.Args[1] == "build":
		cielrouter = cielbuild
	case len(os.Args) == 1+0:
		fallthrough
	default:
		cielhelp()
		return
	}

	fs, err := ci.InitFilesystem("buildkit") // FIXME: configurability
	if err != nil {
		log.Panicln("fs", err)
	}
	defer func() {
		if err := fs.Shutdown(); err != nil {
			log.Panicln("shutdownfs", err)
		}
	}()

	container, err := ci.NewContainer(fs, MachineName)
	if err != nil {
		log.Panicln("container", err)
	}
	defer func() {
		if err := container.Shutdown(); err != nil {
			log.Panicln("shutdown", err)
		}
	}()

	args := []string{}
	if len(os.Args) >= 1+2 {
		args = os.Args[2:]
	}
	if err := cielrouter(container, args); err != nil {
		log.Panicln(err)
	}
}

func cielhelp() {
	fmt.Println(`
commands:
	shell       shell in container
	run <cmd>   run command in container
	build <pkg> build package in container
	(default)   show this information
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
	arg := []string{ACBSPath, "--clear"}
	arg = append(arg, args...)
	if err := cielrun(container, arg); err != nil {
		return err
	}
	// TODO: pick up the package; collect acbs-build log, autobuild log ...
	// TODO: handling multi-package building
	if err := collect(container.FS); err != nil {
		return err
	}
	return nil
}

func collect(fs *ci.ContainerFilesystem) error {
	const CollectedDir = "collected"
	os.Mkdir(CollectedDir, 0755)
	os.Mkdir(CollectedDir+"/repo", 0755)
	os.Rename(fs.DiffDir("/var/log/apt/history.log"), CollectedDir+"/apt-history.log")
	os.Rename(fs.DiffDir("/var/log/apt/term.log"), CollectedDir+"/apt-term.log")
	os.Rename(fs.DiffDir("/var/log/acbs/acbs-build.log"), CollectedDir+"/acbs-build.log")
	os.Rename(fs.DiffDir(findBuildLog(CollectedDir+"/acbs-build.log")), CollectedDir+"/autobuild.log")
	targets := []string{"amd64", "noarch"}
	for _, target := range targets {
		log.Println(os.Rename(fs.DiffDir("/os-"+target), CollectedDir+"/repo/os-"+target))
	}
	return nil
}

func findBuildLog(path string) string {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return ""
	}
	s := string(b)
	const prefix = " Build log: "
	const suffix = "\n"
	prefixindex := strings.LastIndex(s, prefix)
	if prefixindex == -1 {
		return ""
	}
	s = s[prefixindex+len(prefix):]
	suffixindex := strings.Index(s, suffix)
	if suffixindex == -1 {
		return ""
	}
	s = s[:suffixindex]
	return s
}
