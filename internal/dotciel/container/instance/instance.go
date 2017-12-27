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

	"ciel/internal/abstract"
	"ciel/internal/dotciel/container/filesystem"
	"ciel/internal/dotciel/container/nspawn"
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

func (i *Instance) InstCreateFileSystem() error {
	layersDir := path.Join(i.InstSubDir(), LayerDirName)
	return overlayfs.Create(layersDir)
}
func (i *Instance) InstFileSystem() filesystem.FileSystem {
	inst := overlayfs.FromPath(i.Parent.DistDir(), path.Join(i.InstSubDir(), LayerDirName))
	inst.MountPoint = "./" + i.Name
	return inst
}

func (i *Instance) InstMount() error {
	ofs := i.InstFileSystem()
	if !utils.Lock(i.InstLockFile()) {
		if i.InstMounted() {
			return ErrLock
		}
		utils.Unlock(i.InstLockFile())
	}
	if err := ofs.Mount(false); err != nil {
		utils.Unlock(i.InstLockFile())
		return err
	}
	return nil
}
func (i *Instance) InstUnmount() error {
	i.InstUnlockContainer()
	ofs := i.InstFileSystem()
	var err error
	if i.InstMounted() {
		if err := ofs.Unmount(); err != nil {
			return err
		}
	} else {
		err = os.ErrNotExist
	}
	os.Remove(i.InstMountPoint())
	utils.Unlock(i.InstLockFile())
	return err
}

func (i *Instance) InstMounted() bool {
	a, err := ioutil.ReadFile("/proc/self/mountinfo")
	s := string(a)
	list := strings.Split(s, "\n")
	match, _ := filepath.Abs(i.InstMountPoint())
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

func (i *Instance) InstRun(ctx context.Context, boot bool, network bool, containerArgs []string, args ...string) (int, error) {
	defer RecoverTerminalAttr()
	machineId := i.Name + "_" + utils.RandomString(5)

	if !i.InstRefractoryPeriodEnter() {
		return -1, ErrRefractory
	}

	println(i.InstSubDir())

	if boot && nspawn.IsBootable(i.InstMountPoint()) {
		toBoot := false
		if oldMachineId := i.InstMachineId(); oldMachineId == "" {
			i.InstBootedEnter()
			i.InstSetMachineId(machineId)
			toBoot = true
		} else if !nspawn.MachineRunning(nspawn.MachineStatus(ctx, oldMachineId)) {
			i.InstUnlockContainer()
			i.InstBootedEnter()
			i.InstSetMachineId(machineId)
			toBoot = true
		} else {
			machineId = oldMachineId
		}
		i.InstRefractoryPeriodLeave()
		if toBoot {
			if network {
				containerArgs = append([]string{"--network-zone=ciel"}, containerArgs...)
			}
			_, err := nspawn.SystemdNspawn(ctx, i.InstMountPoint(), true, machineId, containerArgs...)

			// NOTE: This will be blocked until the container completely booted up.
			if _, ok := err.(nspawn.ErrCancelled); ok {
				i.InstUnsetMachineId()
				i.InstBootedLeave()
				return -1, err
			}
		}
		defer func() {
			if !i.InstRunning() {
				i.InstUnsetMachineId()
				i.InstBootedLeave()
			}
		}()
		return nspawn.SystemdRun(ctx, machineId, args...)
	}

	i.InstSetMachineId(machineId)

	defer i.InstUnsetMachineId()
	defer i.InstRefractoryPeriodLeave()

	return nspawn.SystemdNspawn(ctx, i.InstMountPoint(), false, machineId, args...)
}

func (i *Instance) InstStop(ctx context.Context) error {
	if i.InstMachineId() == "" {
		return ErrNoMachineId
	}
	var err error
	if i.InstBooted() {
		err = nspawn.MachinectlPoweroff(ctx, i.InstMachineId())
		if err == nil {
			i.InstBootedLeave()
			i.InstUnsetMachineId()
		}
	} else {
		err = nspawn.MachinectlTerminate(ctx, i.InstMachineId())
		if err == nil {
			i.InstUnsetMachineId()
			i.InstRefractoryPeriodLeave()
		}
	}
	return err
}

const cRST = "\x1b[0m"
const cGRE = "\x1b[92m"
const cGRY = "\x1b[37m"
const cRED = "\x1b[91m"
const cCYA = "\x1b[96m"

func (i *Instance) InstFileSystemStat() string {
	if i.InstMounted() {
		return cGRE + "mounted" + cRST
	}
	return cRST + "free" + cRST
}

func (i *Instance) InstContainerRunningStat() (status, boot string) {
	if utils.Locked(i.InstRefractoryFile()) {
		return cRED + "locked" + cRST, cGRY + "unknown" + cRST
	}
	if i.InstMachineId() != "" {
		if i.InstBooted() {
			status := nspawn.MachineStatus(context.Background(), i.InstMachineId())
			if !nspawn.MachineRunning(status) {
				if nspawn.MachineDead(status) {
					i.InstBootedLeave()
					i.InstUnsetMachineId()
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

func (i *Instance) InstRunning() bool {
	if utils.Locked(i.InstRefractoryFile()) {
		return true
	}
	if i.InstMachineId() != "" {
		if i.InstBooted() {
			status := nspawn.MachineStatus(context.Background(), i.InstMachineId())
			return nspawn.MachineRunning(status) || !nspawn.MachineDead(status)
		}
		return true
	}
	return false
}

func (i *Instance) InstLockFile() string {
	return path.Join(i.InstSubDir(), LockFileName)
}

func (i *Instance) InstBootedFile() string {
	return path.Join(i.InstSubDir(), BootedFileName)
}

func (i *Instance) InstBootedEnter() bool {
	return utils.Lock(i.InstBootedFile())
}

func (i *Instance) InstBooted() bool {
	return utils.Locked(i.InstBootedFile())
}

func (i *Instance) InstBootedLeave() {
	utils.Unlock(i.InstBootedFile())
}

func (i *Instance) InstRefractoryFile() string {
	return path.Join(i.InstSubDir(), RefractoryFileName)
}

func (i *Instance) InstRefractoryPeriodEnter() bool {
	return utils.Lock(i.InstRefractoryFile())
}

func (i *Instance) InstRefractoryPeriodLeave() {
	utils.Unlock(i.InstRefractoryFile())
}

func (i *Instance) InstMachineIdFile() string {
	return path.Join(i.InstSubDir(), MachineIdFileName)
}

func (i *Instance) InstSubDir() string {
	return path.Join(i.BasePath, i.Name)
}

func (i *Instance) InstMachineId() string {
	b, _ := ioutil.ReadFile(i.InstMachineIdFile())
	return strings.TrimSpace(string(b))
}

func (i *Instance) InstSetMachineId(machineId string) error {
	return ioutil.WriteFile(i.InstMachineIdFile(), []byte(machineId), 0644)
}

func (i *Instance) InstUnsetMachineId() error {
	return os.Remove(i.InstMachineIdFile())
}

func (i *Instance) InstMountPoint() string {
	return path.Join(i.Parent.GetCiel().GetBasePath(), i.Name)
}

func (i *Instance) InstUnlockFileSystem() {
	i.InstUnmount()
	os.Remove(i.InstLockFile())
	os.Remove(i.InstMountPoint())
}

func (i *Instance) InstUnlockContainer() {
	i.InstStop(context.Background())
	os.Remove(i.InstBootedFile())
	os.Remove(i.InstMachineIdFile())
	os.Remove(i.InstRefractoryFile())

}

func (i *Instance) InstShellPath(user string) (string, error) {
	shell := "/bin/sh"
	passwdFileName := path.Join(i.InstMountPoint(), "/etc/passwd")
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
