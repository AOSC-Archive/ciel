package main

import (
	"log"
	"os/exec"

	"ciel/internal/cieldir.1"
)

func Untar(i *cieldir.CielDir, tar string) {
	cmd := exec.Command("tar", "-xf", tar, "-C", i.DistDir())
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(string(output))
	}
}
