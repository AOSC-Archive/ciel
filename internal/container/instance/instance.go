package instance

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	"ciel/display"
	"ciel/internal/abstract"
	"ciel/internal/container/filesystem"
	"ciel/ipc"
	"ciel/overlayfs"
	"ciel/proc-api"
	"ciel/systemd-api/machined"
	"ciel/systemd-api/nspawn"
)

const (
	LayerDirName         = "layers"
	SemIdFileSystemMutex = 0x11
	SemIdRunMutex        = 0x22
)

var (
	ErrMode = errors.New("another instance is running in exclusive mode")
)

type Instance struct {
	Parent   abstract.Container
	BasePath string
	Name     string
}

func (i *Instance) Init() error {
	layersDir := path.Join(i.Dir(), LayerDirName)
	return overlayfs.Create(layersDir)
}
func (i *Instance) FileSystem() filesystem.FileSystem {
	inst := overlayfs.FromPath(i.Parent.DistDir(), path.Join(i.Dir(), LayerDirName))
	inst.MountPoint = "./" + i.Name
	return inst
}

func (i *Instance) MountPoint() string {
	return path.Join(i.Parent.GetCiel().GetBasePath(), i.Name)
}
func (i *Instance) MountLocal() error {
	fs := i.FileSystem()
	CriticalSection := i.FileSystemLock()

	CriticalSection.Lock()
	defer CriticalSection.Unlock()

	if !i.Mounted() {
		if err := fs.MountLocal(); err != nil {
			return err
		}
	}
	return nil
}
func (i *Instance) Mount() error {
	fs := i.FileSystem()
	CriticalSection := i.FileSystemLock()

	CriticalSection.Lock()
	defer CriticalSection.Unlock()

	if !i.Mounted() {
		if err := fs.Mount(false); err != nil {
			return err
		}
		i.Parent.GetCiel().GetTree().MountHandler(i, true)
	}
	return nil
}
func (i *Instance) Unmount() error {
	i.Stop(context.Background())
	fs := i.FileSystem()
	CriticalSection := i.FileSystemLock()

	CriticalSection.Lock()
	defer CriticalSection.Unlock()

	var err error
	if i.Mounted() {
		i.Parent.GetCiel().GetTree().MountHandler(i, false)
		d.ITEM("unmount " + i.Name)
		if err := fs.Unmount(); err != nil {
			d.FAILED_BECAUSE(err.Error())
			return err
		}
		d.OK()
	} else {
		d.ITEM("unmount " + i.Name)
		d.SKIPPED()
		err = os.ErrNotExist
	}
	d.ITEM("remove mount point")
	tryRemove(i.MountPoint())
	return err
}

func tryRemove(path string) {
	err := os.Remove(path)
	if err == nil {
		d.OK()
		return
	}
	if os.IsNotExist(err) {
		d.SKIPPED()
		return
	}
	d.FAILED_BECAUSE(err.Error())
}
func (i *Instance) Mounted() bool {
	return proc.Mounted(i.MountPoint())
}

func (i *Instance) Run(ctx context.Context, ctnInfo *nspawn.ContainerInfo, runInfo *nspawn.RunInfo) (int, error) {
	defer RecoverTerminalAttr()
	machineId := fmt.Sprintf("%s_%x", i.Name, ipc.GenFileKey(i.Parent.GetCiel().GetBasePath(), 0))

	CriticalSection := i.RunLock()

	if i.RunningAsExclusiveMode() {
		return -1, ErrMode
	}

	var boot bool

	if ctnInfo.Init && nspawn.IsBootable(i.MountPoint()) {
		boot = true
	} else {
		boot = false
	}

	if boot {
		var err error
		CriticalSection.Lock()
		if !i.Running() {
			err = nspawn.SystemdNspawnBoot(ctx, machineId, i.MountPoint(), ctnInfo)
		}
		CriticalSection.Unlock()
		if err != nil {
			return -1, err
		}
		if !i.RunningAsBootMode() {
			return -1, ErrMode
		}
		if runInfo.UseSystemdRun {
			return nspawn.SystemdRun(ctx, machineId, runInfo)
		}
		return nspawn.MachinectlShell(ctx, machineId, runInfo)
	} else {
		return nspawn.SystemdNspawnRun(ctx, machineId, i.MountPoint(), ctnInfo, runInfo)
	}
}

func (i *Instance) Stop(ctx context.Context) error {
	d.ITEM("stop " + i.Name)
	if !i.Running() {
		d.SKIPPED()
		return nil
	}
	var err error
	if i.RunningAsBootMode() {
		err = nspawn.MachinectlPowerOff(ctx, i.MachineId())
	} else {
		err = nspawn.MachinectlTerminate(ctx, i.MachineId())
	}
	d.ERR(err)
	return err
}

func (i *Instance) Running() bool {
	m := machined.NewManager()
	_, err := m.GetMachine(i.MachineId())
	return err == nil
}

func (i *Instance) RunningAsBootMode() bool {
	m := machined.NewManager()
	machine, err := m.GetMachine(i.MachineId())
	if err != nil {
		return false
	}
	leader, err := machine.Leader()
	if err != nil {
		log.Fatalln(err)
	}
	host, err := proc.GetParentProcessID(leader)
	if err != nil {
		log.Fatalln(err)
	}
	cmdline, err := proc.GetCommandLineByPID(host)
	if err != nil {
		log.Fatalln(err)
	}
	for _, arg := range cmdline {
		if arg == "-b" || arg == "--boot" {
			return true
		}
	}
	return false
}

func (i *Instance) RunningAsExclusiveMode() bool {
	m := machined.NewManager()
	machine, err := m.GetMachine(i.MachineId())
	if err != nil {
		return false
	}
	leader, err := machine.Leader()
	if err != nil {
		log.Fatalln(err)
	}
	host, err := proc.GetParentProcessID(leader)
	if err != nil {
		log.Fatalln(err)
	}
	cmdline, err := proc.GetCommandLineByPID(host)
	if err != nil {
		log.Fatalln(err)
	}
	for _, arg := range cmdline {
		if arg == "-b" || arg == "--boot" {
			return false
		}
	}
	return true
}

func (i *Instance) Dir() string {
	return path.Join(i.BasePath, i.Name)
}

func (i *Instance) MachineId() string {
	return fmt.Sprintf("%s_%x", i.Name, ipc.GenFileKey(i.Parent.GetCiel().GetBasePath(), 0))
}

func (i *Instance) FileSystemLock() ipc.Mutex {
	return ipc.NewMutex(i.Dir(), SemIdFileSystemMutex, true)
}

func (i *Instance) RunLock() ipc.Mutex {
	return ipc.NewMutex(i.Dir(), SemIdRunMutex, true)
}

func (i *Instance) Shell(user string) (string, error) {
	shell := "/bin/sh"
	passwdFileName := path.Join(i.MountPoint(), "/etc/passwd")
	a, err := ioutil.ReadFile(passwdFileName)
	if err != nil {
		return "", err
	}
	passwd := string(a)
	for _, userInfo := range strings.Split(passwd, "\n") {
		if userInfo == "" {
			continue
		}
		fields := strings.Split(userInfo, ":")
		if fields[0] == user {
			shell = fields[6]
		}
	}
	return shell, nil
}

func (i *Instance) GetContainer() abstract.Container { return i.Parent }
