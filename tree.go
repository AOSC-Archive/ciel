package main

import (
	"flag"
	"os"

	"ciel/internal/ciel"
)

func clone() {
	basePath := flagCielDir()
	parse()

	i := &ciel.Ciel{BasePath: *basePath}
	i.Check()
	t := i.Tree()

	os.Exit(t.Clone(flag.Arg(0)))
}

func pull() {
	basePath := flagCielDir()
	parse()

	i := &ciel.Ciel{BasePath: *basePath}
	i.Check()
	t := i.Tree()

	os.Exit(t.Pull())
}
