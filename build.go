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

	// >> /etc/acbs/forest.conf
	// [default]
	// location = /var/lib/acbs/repo

	// >> /usr/lib/autobuild3/etc/autobuild/ab3cfg.sh
	// #!/bin/bash
	// ##Autobuild user config override
	// # See a list of options in ab3_defcfg.sh.
	// ABMPM=dpkg  # Your main PM
	// ABAPMS=  # Other PMs
	// MTER="Null Packager <null@aosc.xyz>"
	// ABINSTALL=

	// TODO: interactive configuring
}
