package main

import (
	"io"
	"os"
	"os/exec"
)

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
