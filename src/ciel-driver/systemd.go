package ciel

import (
	"errors"
	"os"
	"os/exec"
	"syscall"
	"time"
)

func (c *Container) systemdNspawnBoot() {
	c.fs.lock.RLock()
	cmd := exec.Command("/usr/bin/systemd-nspawn",
		"--quiet",
		"--boot",
		// "--property=CPUQuota=80%", // FIXME: configurability
		// "--property=MemoryMax=70%",
		// "--property=MemoryHigh=60%",
		// "--property=MemoryLow=40%",
		"-M", c.name,
		"-D", c.fs.target,
	)
	c.fs.lock.RUnlock()
	if err := cmd.Start(); err != nil {
		panic(err)
	}
	go func() {
		if err := cmd.Wait(); err != nil {
			c.lock.Lock()
			if c.active {
				c.active = false
				close(c.cancel)
				c.cancel = make(chan struct{})
			}
			c.lock.Unlock()
		}
	}()
	for !c.isSystemRunning() {
		select {
		case <-c.cancel:
			panic("container dead")
		default:
			time.Sleep(time.Millisecond * 100)
		}
	}
	c.lock.Lock()
	c.active = true
	c.lock.Unlock()
}

func (c *Container) isSystemRunning() bool {
	return exec.Command("/usr/bin/systemctl", "is-system-running", "-M", c.name).Run() == nil
}

func (c *Container) isSystemShutdown() bool {
	return exec.Command("/usr/bin/machinectl", "status", c.name).Run() != nil
}

func (c *Container) machinectlPoweroff() error {
	c.lock.Lock()
	defer c.lock.Unlock()

	cmd := exec.Command("/usr/bin/machinectl", "poweroff", c.name)
	b, err := cmd.CombinedOutput()
	c.active = false
	close(c.cancel)
	c.cancel = make(chan struct{})
	for !c.isSystemShutdown() {
		select {
		case <-c.cancel:
			panic("container dead")
		default:
			time.Sleep(time.Millisecond * 100)
		}
	}

	if err != nil {
		return errors.New(string(b))
	}
	return nil
}

func (c *Container) systemdRun(proc string, args ...string) int {
	c.lock.RLock()
	a := c.active
	c.lock.RUnlock()
	if !a {
		return -1
	}
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
	if err != nil {
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
