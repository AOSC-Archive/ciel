package instance

import (
	"context"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"ciel/internal/ciel/abstract"
	"ciel/internal/ciel/container/filesystem"
	"ciel/internal/ciel/container/nspawn"
	"ciel/internal/display"
	"ciel/internal/overlayfs"
	"ciel/internal/utils"
)

const (
	LayerDirName       = "layers"
	LockFileName       = "lock"
	RefractoryFileName = "refractory"
	BootedFileName     = "booted"
	MachineIdFileName  = "machineid"
)

var (
	ErrLock        = errors.New("failed to require the lock")
	ErrNoMachineId = errors.New("no machineId")
	ErrRefractory  = errors.New("another instance is running")
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
func (i *Instance) Mount() error {
	ofs := i.FileSystem()
	if !utils.Lock(i.FileSystemLock()) {
		if i.Mounted() {
			return ErrLock
		}
		os.Remove(i.FileSystemLock())
	}
	if err := ofs.Mount(false); err != nil {
		os.Remove(i.FileSystemLock())
		return err
	}
	return nil
}
func (i *Instance) Unmount() error {
	i.Stop(context.Background())
	ofs := i.FileSystem()
	var err error
	d.ITEM("unmount")
	if i.Mounted() {
		if err := ofs.Unmount(); err != nil {
			d.FAILED_BECAUSE(err.Error())
			return err
		}
		d.OK()
	} else {
		d.SKIPPED()
		err = os.ErrNotExist
	}
	d.ITEM("remove mount point")
	tryRemove(i.MountPoint())
	d.ITEM("remove lock")
	tryRemove(i.FileSystemLock())
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
	machineId := i.Name + "_" + utils.RandomString(5)

	if !i.EnterRefractory() {
		return -1, ErrRefractory
	}

	if boot && nspawn.IsBootable(i.MountPoint()) {
		toBoot := false
		if oldMachineId := i.MachineId(); oldMachineId == "" {
			i.EnterBoot()
			i.SetMachineId(machineId)
			toBoot = true
		} else if !nspawn.MachineRunning(nspawn.MachineStatus(ctx, oldMachineId)) {
			i.Stop(context.Background())
			i.EnterBoot()
			i.SetMachineId(machineId)
			toBoot = true
		} else {
			machineId = oldMachineId
		}
		i.LeaveRefractory()
		if toBoot {
			if network {
				containerArgs = append([]string{"--network-zone=ciel"}, containerArgs...)
			}
			_, err := nspawn.SystemdNspawn(ctx, i.MountPoint(), true, machineId, containerArgs...)

			// NOTE: This will be blocked until the container completely booted up.
			if _, ok := err.(nspawn.ErrCancelled); ok {
				i.UnsetMachineId()
				i.LeaveBoot()
				return -1, err
			}
		}
		defer func() {
			if !i.Running() {
				i.UnsetMachineId()
				i.LeaveBoot()
			}
		}()
		return nspawn.SystemdRun(ctx, machineId, args...)
	}

	i.SetMachineId(machineId)

	defer i.UnsetMachineId()
	defer i.LeaveRefractory()

	return nspawn.SystemdNspawn(ctx, i.MountPoint(), false, machineId, args...)
}

func (i *Instance) Stop(ctx context.Context) error {
	d.ITEM("stop")
	if i.MachineId() == "" {
		d.SKIPPED()
		return ErrNoMachineId
	}
	var err error
	if i.Booted() {
		err = nspawn.MachinectlPoweroff(ctx, i.MachineId())
	} else {
		err = nspawn.MachinectlTerminate(ctx, i.MachineId())
	}
	d.ERR(err)
	if err == nil {
		d.ITEM("remove boot flag")
		tryRemove(i.BootLock())
		d.ITEM("remove machine id")
		tryRemove(i.MachineIdFile())
		d.ITEM("remove boot lock")
		tryRemove(i.RefractoryLock())
	}
	return err
}

const cRST = "\x1b[0m"
const cGRE = "\x1b[92m"
const cGRY = "\x1b[37m"
const cRED = "\x1b[91m"
const cCYA = "\x1b[96m"

func (i *Instance) FileSystemStat() string {
	if i.Mounted() {
		return cGRE + "mounted" + cRST
	}
	return cRST + "free" + cRST
}

func (i *Instance) ContainerStat() (status, boot string) {
	if utils.Locked(i.RefractoryLock()) {
		return cRED + "locked" + cRST, cGRY + "unknown" + cRST
	}
	if i.MachineId() != "" {
		if i.Booted() {
			status := nspawn.MachineStatus(context.Background(), i.MachineId())
			if !nspawn.MachineRunning(status) {
				if nspawn.MachineDead(status) {
					i.LeaveBoot()
					i.UnsetMachineId()
					return cGRY + "offline" + cRST, cGRY + "no" + cRST
				}
				return cGRY + "linger" + cRST, cCYA + "yes" + cRST
			}
			return cCYA + "running" + cRST, cCYA + "yes" + cRST
		}
		return cCYA + "running" + cRST, cGRY + "no" + cRST
	}
	return cGRY + "offline" + cRST, cGRY + "no" + cRST
}

func (i *Instance) Running() bool {
	if utils.Locked(i.RefractoryLock()) {
		return true
	}
	if i.MachineId() != "" {
		if i.Booted() {
			status := nspawn.MachineStatus(context.Background(), i.MachineId())
			return nspawn.MachineRunning(status) || !nspawn.MachineDead(status)
		}
		return true
	}
	return false
}

func (i *Instance) BootLock() string {
	return path.Join(i.Dir(), BootedFileName)
}
func (i *Instance) EnterBoot() bool {
	return utils.Lock(i.BootLock())
}
func (i *Instance) LeaveBoot() {
	os.Remove(i.BootLock())
}
func (i *Instance) Booted() bool {
	return utils.Locked(i.BootLock())
}

func (i *Instance) RefractoryLock() string {
	return path.Join(i.Dir(), RefractoryFileName)
}
func (i *Instance) EnterRefractory() bool {
	return utils.Lock(i.RefractoryLock())
}
func (i *Instance) LeaveRefractory() {
	os.Remove(i.RefractoryLock())
}

func (i *Instance) Dir() string {
	return path.Join(i.BasePath, i.Name)
}

func (i *Instance) MachineIdFile() string {
	return path.Join(i.Dir(), MachineIdFileName)
}
func (i *Instance) MachineId() string {
	b, _ := ioutil.ReadFile(i.MachineIdFile())
	return strings.TrimSpace(string(b))
}
func (i *Instance) SetMachineId(machineId string) error {
	return ioutil.WriteFile(i.MachineIdFile(), []byte(machineId), 0644)
}
func (i *Instance) UnsetMachineId() error {
	return os.Remove(i.MachineIdFile())
}

func (i *Instance) FileSystemLock() string {
	return path.Join(i.Dir(), LockFileName)
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
