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
	"ciel/internal/overlayfs"
	"ciel/internal/utils"
)

const LayerDirName = "layers"
const LockFileName = "lock"
const RefractoryFileName = "refractory"
const BootedFileName = "booted"
const MachineIdFileName = "machineid"

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
		utils.Unlock(i.FileSystemLock())
	}
	if err := ofs.Mount(false); err != nil {
		utils.Unlock(i.FileSystemLock())
		return err
	}
	return nil
}
func (i *Instance) Unmount() error {
	i.UnlockContainer()
	ofs := i.FileSystem()
	var err error
	if i.Mounted() {
		if err := ofs.Unmount(); err != nil {
			return err
		}
	} else {
		err = os.ErrNotExist
	}
	os.Remove(i.MountPoint())
	utils.Unlock(i.FileSystemLock())
	return err
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

	println(i.Dir())

	if boot && nspawn.IsBootable(i.MountPoint()) {
		toBoot := false
		if oldMachineId := i.MachineId(); oldMachineId == "" {
			i.EnterBoot()
			i.SetMachineId(machineId)
			toBoot = true
		} else if !nspawn.MachineRunning(nspawn.MachineStatus(ctx, oldMachineId)) {
			i.UnlockContainer()
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
	if i.MachineId() == "" {
		return ErrNoMachineId
	}
	var err error
	if i.Booted() {
		err = nspawn.MachinectlPoweroff(ctx, i.MachineId())
		if err == nil {
			i.LeaveBoot()
			i.UnsetMachineId()
		}
	} else {
		err = nspawn.MachinectlTerminate(ctx, i.MachineId())
		if err == nil {
			i.UnsetMachineId()
			i.LeaveRefractory()
		}
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
	utils.Unlock(i.BootLock())
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
	utils.Unlock(i.RefractoryLock())
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
func (i *Instance) UnlockFileSystem() {
	i.Unmount()
	os.Remove(i.FileSystemLock())
	os.Remove(i.MountPoint())
}

func (i *Instance) UnlockContainer() {
	i.Stop(context.Background())
	os.Remove(i.BootLock())
	os.Remove(i.MachineIdFile())
	os.Remove(i.RefractoryLock())

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
