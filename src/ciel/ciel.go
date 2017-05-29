package main

import (
	"ci"
	"log"
)

const MachineName = "buildkit"
const ACBSPath = "/bin/acbs-build"

func main() {
	fs, err := ci.InitFilesystem("/root/ciel/buildkit")
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

	cmd := container.Exec(ACBSPath, "-c", "nano")
	redirectStdOutErr(cmd)
	if err := cmd.Run(); err != nil {
		log.Println("acbs-build", err)
	}
}
