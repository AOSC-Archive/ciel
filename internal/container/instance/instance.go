package instance

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
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
	SemIdFileSystemMutex = 101
	SemIdBootMutex       = 1
)

var (
	ErrMode = errors.New("another instance is running in excluded mode")
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
	a, err := ioutil.ReadFile("/proc/self/mountinfo")
	s := string(a)
	list := strings.Split(s, "\n")
	match, _ := filepath.Abs(i.MountPoint())
	for _, item := range list {
		if item == "" {
			continue
		}
		fields := strings.Split(item, " ")
		if fields[4] == match {
			return true
		}
	}
	if err != nil {
		log.Panicln(err)
	}
	return false
}

func (i *Instance) Run(ctx context.Context, boot bool, network bool, containerArgs []string, args ...string) (int, error) {
	defer RecoverTerminalAttr()
	machineId := fmt.Sprintf("%s_%x", i.Name, ipc.GenFileKey(i.Parent.GetCiel().GetBasePath(), 0))

	CriticalSection := ipc.NewMutex(i.MountPoint(), SemIdBootMutex, true)

	if i.RunningAsExcludedMode() {
		return -1, ErrMode
	}

	var preferBoot bool

	if boot && nspawn.IsBootable(i.MountPoint()) {
		preferBoot = true
	} else {
		preferBoot = false
	}

	if preferBoot {
		var err error
		CriticalSection.Lock()
		if !i.Running() {
			if network {
				containerArgs = append([]string{"--network-zone=ciel"}, containerArgs...)
			}
			_, err = nspawn.SystemdNspawnBoot(ctx, i.MountPoint(), machineId, containerArgs...)
		}
		CriticalSection.Unlock()
		if !i.RunningAsBootMode() {
			return -1, ErrMode
		}
		if err != nil {
			return -1, err
		}
		return nspawn.SystemdRun(ctx, machineId, args...)
	} else {
		es, err := nspawn.SystemdNspawnRun(ctx, i.MountPoint(), machineId, args...)
		return es, err
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
		err = nspawn.MachinectlPoweroff(ctx, i.MachineId())
	} else {
		err = nspawn.MachinectlTerminate(ctx, i.MachineId())
	}
	d.ERR(err)
	return err
}

const cRST = "\x1b[0m"
const cGRE = "\x1b[92m"
const cGRY = "\x1b[37m"
const cCYA = "\x1b[96m"

func (i *Instance) FileSystemStat() string {
	if i.Mounted() {
		return cGRE + "mounted" + cRST
	}
	return cRST + "free" + cRST
}

func (i *Instance) ContainerStat() (status, boot string) {
	//if utils.Locked(i.RefractoryLock()) {
	//	return cRED + "locked" + cRST, cGRY + "unknown" + cRST
	//}
	if i.Running() {
		status = cCYA + "running" + cRST
	} else {
		status = cGRY + "offline" + cRST
	}
	if i.RunningAsBootMode() {
		boot = cCYA + "yes" + cRST
	} else {
		boot = cGRY + "no" + cRST
	}
	return
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

func (i *Instance) RunningAsExcludedMode() bool {
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
	return ipc.NewMutex(i.MountPoint(), SemIdFileSystemMutex, true)
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
