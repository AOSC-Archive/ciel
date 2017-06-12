package ciel

import (
	"os"
	"os/exec"
	"syscall"
)

func (c *Container) systemdNspawnBoot() {
	c.fs.lock.RLock()
	cmd := exec.Command("/usr/bin/systemd-nspawn",
		"--quiet",
		"--boot",
		"--property=CPUQuota=80%", // FIXME: configurability
		"--property=MemoryMax=70%",
		"--property=MemoryHigh=60%",
		"--property=MemoryLow=40%",
		"-M", c.name,
		"-D", c.fs.target,
	)
	c.fs.lock.RUnlock()
	if err := cmd.Start(); err != nil {
		panic(err)
	}
}

func (c *Container) machinectlPoweroff() error {
	cmd := exec.Command("/usr/bin/machinectl", "poweroff", c.name)
	return cmd.Run()
}

func (c *Container) systemdRun(proc string, args ...string) int {
	// TODO: implement systemdRun()
	subArgs := append([]string{proc}, args...)
	subArgs = append([]string{
		"--quiet",
		"--wait",
		"--pty",
		"-M", c.name,
	}, subArgs...)
	return cmd("/usr/bin/systemd-run", subArgs...)
}

func (c *Container) systemdNspawnRun(proc string, args ...string) int {
	subArgs := append([]string{proc}, args...)
	c.fs.lock.RLock()
	subArgs = append([]string{
		"--quiet",
		"-M", c.name,
		"-D", c.fs.target,
	}, subArgs...)
	c.fs.lock.RUnlock()
	return cmd("/usr/bin/systemd-nspawn", subArgs...)
}

func cmd(proc string, args ...string) int {
	cmd := exec.Command(proc, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err == nil {
		panic(err)
	}
	err = cmd.Wait()
	if err == nil {
		return 0
	}
	if exitError, ok := err.(*exec.ExitError); ok {
		exitStatus := exitError.Sys().(syscall.WaitStatus)
		return exitStatus.ExitStatus()
	}
	panic(err)
}
