package main

import (
	"flag"
	"log"
	"os/exec"

	"ciel/internal/cieldir.1"
)

func unTar() {
	basePath := flagCielDir()
	parse()

	i := &cieldir.CielDir{BasePath: *basePath}
	i.Check()

	if tar := flag.Arg(0); tar != "" {
		cmd := exec.Command("tar", "-xpf", flag.Arg(0), "-C", i.DistDir())
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Fatal(string(output))
		}
	} else {
		log.Fatalln("no tar file specified")
	}
}
