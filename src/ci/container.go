package ci

import (
	"log"
	"os"
	"os/exec"
)

type ContainerInstance struct {
	Name string
	FS   *ContainerFilesystem

	NoBooting bool
	Cmd       string
	Args      []string

	Wait chan struct{}
}

func NewContainer(fs *ContainerFilesystem, machine string) *ContainerInstance {
	container := &ContainerInstance{Name: machine, FS: fs}
	return container
}

func (c *ContainerInstance) Startup() error {
	if !c.NoBooting {
		return c.startupBoot()
	} else {
		return c.startupChroot()
	}
}

func (c *ContainerInstance) startupBoot() error {
	args := []string{
		"--quiet",
		"--boot",
		"--property=CPUQuota=80%", // FIXME: configurability
		"--property=MemoryMax=70%",
		"--property=MemoryHigh=60%",
		"--property=MemoryLow=40%",
		"-M", c.Name,
		"-D", c.FS.Target,
	}
	cmd := exec.Command("/usr/bin/systemd-nspawn", args...)
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil { // Create and boot the container
		return err
	}
	wait := make(chan struct{}) // Exit signal channel
	c.Wait = wait
	go func() {
		if err := cmd.Wait(); err != nil {
			defer c.FS.Shutdown()
			log.Panic(err) // systemd-nspawn exited with non-zero exit code
		}
		close(wait)
	}()
	for !c.IsAlive() { // Wait for booting
	}
	return nil
}

func (c *ContainerInstance) startupChroot() error {
	args := []string{
		"--quiet",
		"-M", c.Name,
		"-D", c.FS.Target,
	}
	args = append(args, c.Cmd)
	args = append(args, c.Args...)
	cmd := exec.Command("/usr/bin/systemd-nspawn", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (c *ContainerInstance) Shutdown() error {
	if c.NoBooting {
		return nil
	}
	cmd := exec.Command("/usr/bin/machinectl", "poweroff", c.Name)
	if err := cmd.Run(); err != nil {
		return err
	}
	<-c.Wait // Wait for systemd-nspawn
	return nil
}

func (c *ContainerInstance) IsAlive() bool {
	cmd := c.Exec("/bin/ls", "/root")
	_, err := cmd.CombinedOutput()
	return err == nil
}

func (c *ContainerInstance) Exec(arg ...string) *exec.Cmd {
	arghead := []string{"shell", "-q", "root@" + c.Name}
	arg = append(arghead, arg...)
	return exec.Command("/usr/bin/machinectl", arg...)
}
