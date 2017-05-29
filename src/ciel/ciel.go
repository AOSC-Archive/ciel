package main

import (
	"ci"
	"io"
	"log"
	"os"
	"os/exec"
)

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

	container, err := ci.NewContainer(fs, "buildkit")
	if err != nil {
		log.Panicln("container", err)
	}
	defer func() {
		if err := container.Shutdown(); err != nil {
			log.Panicln("shutdown", err)
		}
	}()

	cmd := container.Exec("/bin/acbs-build", "-c", "nano")
	redirectStdOutErr(cmd)
	if err := cmd.Run(); err != nil {
		log.Println("acbs-build", err)
	}
}

func redirectStdOutErr(cmd *exec.Cmd) {
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	pipeIO(stdout, os.Stdout)
	pipeIO(stderr, os.Stderr)
}

func pipeIO(r io.Reader, w io.Writer) {
	go func() {
		for {
			if _, err := io.Copy(w, r); err != nil {
				return
			}
		}
	}()
}
