package ciel

const _SHELLPATH = "/bin/bash"

type Container struct {
	name string
	fs   *filesystem

	BootPreferred bool
	active        bool
}

func New(name, fsPath string) *Container {
	return &Container{
		name:          name,
		fs:            new(filesystem),
		BootPreferred: true,
		active:        false,
	}
}

func (c *Container) Command(cmdline string) int {
	return c.CommandRaw(_SHELLPATH, "-l", "-c", cmdline)
}
func (c *Container) CommandRaw(proc string, args ...string) int {
	if !c.fs.active {
		c.Mount()
	}
	if c.BootPreferred && c.Bootable() {
		if !c.active {
			c.systemdNspawnBoot()
		}
		return c.systemdRun()
	} else {
		return c.systemdNspawnRun()
	}
}

func (c *Container) SetBaseDir(path string) {
	c.fs.setBaseDir(path)
}
func (c *Container) Bootable() bool {
	return c.fs.bootable()
}
func (c *Container) Mount() error {
	return c.fs.mount()
}
func (c *Container) Unmount() error {
	return c.fs.unmount()
}
