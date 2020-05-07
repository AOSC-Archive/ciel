package pkgtree

import (
	"os/exec"

	"log"
	"os"
	"syscall"

	"github.com/AOSC-Dev/ciel/internal/abstract"
)

type Tree struct {
	Parent   abstract.Ciel
	BasePath string
}

func (t *Tree) Clone(remote string) int {
	cmd := exec.Command("git", "clone", remote, t.BasePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if exitErr, ok := err.(*exec.ExitError); ok {
		return exitErr.Sys().(syscall.WaitStatus).ExitStatus()
	}
	if err != nil {
		log.Fatalln(err)
	}
	return 0
}

func (t *Tree) Pull() int {
	cmd := exec.Command("git", "-C", t.BasePath, "pull", "--rebase")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if exitErr, ok := err.(*exec.ExitError); ok {
		return exitErr.Sys().(syscall.WaitStatus).ExitStatus()
	}
	if err != nil {
		log.Fatalln(err)
	}
	return 0
}
