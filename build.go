package main

import (
	"ciel/internal/ciel"
	"ciel/internal/ciel/packaging"
)

func buildConfig() {
	basePath := flagCielDir()
	instName := flagInstance()
	parse()

	i := &ciel.Ciel{BasePath: *basePath}
	i.Check()
	c := i.Container()
	c.CheckInst(*instName)

	inst := c.Instance(*instName)
	inst.Mount()

	packaging.DetectToolChain(inst)

	// TODO: interactive configuring
}
