package ci

import (
	"log"
	"os/exec"
)

type ContainerInstance struct {
	Name string
	FS   *ContainerFilesystem
	Wait chan struct{}
}

func NewContainer(fs *ContainerFilesystem, machine string) (*ContainerInstance, error) {
	cmd := exec.Command("/usr/bin/systemd-nspawn", "-qb", "-M", machine, "-D", fs.TargetDir)
	if err := cmd.Start(); err != nil { // Create and boot the container
		return nil, err
	}

	wait := make(chan struct{}) // Exit signal channel
	go func() {
		if err := cmd.Wait(); err != nil {
			log.Panic(err) // systemd-nspawn exited with non-zero exit code
		}
		close(wait)
	}()

	container := &ContainerInstance{Name: machine, FS: fs, Wait: wait}
	for !container.IsAlive() { // Wait for booting
	}
	return container, nil
}

func (c *ContainerInstance) Shutdown() error {
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
