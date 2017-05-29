package ci

import (
	"encoding/base64"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

type ContainerFilesystem struct {
	Base string

	overlay string
	diff    string
	work    string

	TargetDir string
}

func InitFilesystem(bkdir string) (*ContainerFilesystem, error) {
	fs := &ContainerFilesystem{Base: bkdir}
	fs.overlay = fs.Base + ".overlay"
	fs.diff = fs.Base + ".diff"
	fs.work = fs.Base + ".temp"
	rd := make([]byte, 8)
	if _, err := rand.Read(rd); err != nil {
		return nil, err
	}
	fs.TargetDir = os.TempDir() + "/ciel." + base64.RawURLEncoding.EncodeToString(rd)
	os.Mkdir(fs.diff, 0755)
	os.Mkdir(fs.work, 0755)
	os.Mkdir(fs.TargetDir, 0755)
	if _, err := os.Stat(fs.overlay); os.IsNotExist(err) {
		return fs, mount(fs.TargetDir, fs.diff, fs.work, fs.Base)
	} else {
		return fs, mount(fs.TargetDir, fs.diff, fs.work, fs.overlay, fs.Base)
	}
}

func (fs *ContainerFilesystem) Shutdown() error {
	if err := unmount(fs.TargetDir); err != nil {
		return err
	}
	if err := os.Remove(fs.TargetDir); err != nil {
		return err
	}
	if err := os.RemoveAll(fs.work); err != nil {
		return err
	}
	return nil
}

func mount(path string, upperdir string, workdir string, lowerdirs ...string) error {
	return syscall.Mount("overlay", path, "overlay", 0,
		"lowerdir="+strings.Join(lowerdirs, ":")+",upperdir="+upperdir+",workdir="+workdir)
}

func unmount(path string) error {
	return syscall.Unmount(path, 0)
}

type ContainerInstance struct {
	Name string
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

	container := &ContainerInstance{Name: machine, Wait: wait}
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
