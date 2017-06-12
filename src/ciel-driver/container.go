package ciel

import (
	"sync"
)

const _SHELLPATH = "/bin/bash"

type Container struct {
	lock sync.RWMutex

	name string
	fs   *filesystem

	bootPreferred bool
	active        bool

	cancel chan struct{}
}

// New creates a container descriptor, but it won't start the container immediately.
//
// You may want to call Command() after this.
func New(name, baseDir string) *Container {
	c := &Container{
		name:          name,
		fs:            new(filesystem),
		bootPreferred: true,
		active:        false,
		cancel:        make(chan struct{}),
	}
	c.SetBaseDir(baseDir)
	return c
}

// Command is the most useful function.
// It calls the command line with shell (bash) in container, returns the exit code.
//
// Don't worry about mounting file system, starting container and the mode of booting.
// Please check out CommandRaw() for more details.
//
// NOTE: It calls CommandRaw() internally.
func (c *Container) Command(cmdline string) int {
	return c.CommandRaw(_SHELLPATH, "-l", "-c", cmdline)
}

// CommandRaw runs command in container.
//
// It will mount the root file system and start the container automatically,
// when they are not active. It can also choose boot-mode and chroot-mode automatically.
// You may change this behaviour by SetPreference().
func (c *Container) CommandRaw(proc string, args ...string) (exitCode int) {
	if !c.fs.active {
		if err := c.Mount(); err != nil {
			panic(err)
		}
	}
	if c.bootPreferred && c.IsBootable() {
		if !c.active {
			c.systemdNspawnBoot()
		}
		exitCode = c.systemdRun(proc, args...)
	} else {
		c.lock.Lock()
		c.active = true
		c.lock.Unlock()
		exitCode = c.systemdNspawnRun(proc, args...)
		c.lock.Lock()
		c.active = false
		c.lock.Unlock()
	}
	return
}

// Shutdown the container and unmount file system.
func (c *Container) Shutdown() error {
	return c.machinectlPoweroff()
}

// IsContainerActive returns whether the container is running or not.
func (c *Container) IsContainerActive() bool {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.active
}

// SetPreference changes the preference of container.
//
// <boot>: (default: true) CommandRaw() will boot system on container,
// if the file system is bootable.
// When you set it to "false", CommandRaw() will only chroot,
// even the file system is bootable.
func (c *Container) SetPreference(boot bool) {
	c.lock.Lock()
	c.bootPreferred = boot
	c.lock.Unlock()
}

// IsFileSystemActive returns whether the file system has been mounted or not.
func (c *Container) IsFileSystemActive() bool {
	return c.fs.IsActive()
}

// IsBootable returns whether the file system is bootable or not.
//
// NOTE: The basis of determining is the file /usr/lib/systemd/systemd.
func (c *Container) IsBootable() bool {
	return c.fs.IsBootable()
}

// SetBaseDir sets the base directory for components of the container.
func (c *Container) SetBaseDir(path string) {
	c.fs.SetBaseDir(path)
}

// Mount the file system to a temporary directory.
// It will be called automatically by CommandRaw().
func (c *Container) Mount() error {
	return c.fs.Mount()
}

// Unmount the file system, and cleans the temporary directories.
func (c *Container) Unmount() error {
	return c.fs.Unmount()
}
