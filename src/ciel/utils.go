package main

import (
	"io"
	"os"
	"os/exec"
)

func redirectStdIO(cmd *exec.Cmd) {
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
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
