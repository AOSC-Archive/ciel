package main

import (
	"flag"
	"os"

	"ciel/internal/ciel"
)

const (
	GitAOSCOSABBS = "https://github.com/AOSC-Dev/aosc-os-abbs"
)

func clone() {
	basePath := flagCielDir()
	parse()

	i := &ciel.Ciel{BasePath: *basePath}
	i.Check()
	t := i.Tree()

	tree := flag.Arg(0)
	if tree == "" {
		tree = GitAOSCOSABBS
	}
	os.Exit(t.Clone(tree))
}

func pull() {
	basePath := flagCielDir()
	parse()

	i := &ciel.Ciel{BasePath: *basePath}
	i.Check()
	t := i.Tree()

	os.Exit(t.Pull())
}
