package cieldir

import (
	"context"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"ciel/internal/nspawn"
	"ciel/internal/overlayfs"
	"ciel/internal/utils"
)

const LayerDirName = "layers"
const LockFileName = "lock"
const RefractoryFileName = "refractory"
const BootedFileName = "booted"
const MachineIdFileName = "machineid"

var Layers = []string{
	"local",
	"diff",
}

var (
	ErrLock = errors.New("failed to require the lock")
)

func (i *CielDir) CheckInst(name string) {
	if !i.InstExists(name) {
		log.Fatalln("instance '" + name + "' does not exist")
	}
}

func (i *CielDir) InstAdd(name string) {
	utils.Mkdir(i.InstSubDir(name))
	layersDir := path.Join(i.InstSubDir(name), LayerDirName)
	utils.Mkdir(layersDir)
	for _, dir := range Layers {
		utils.Mkdir(path.Join(layersDir, dir))
	}
}

func (i *CielDir) InstDel(name string) {
	os.RemoveAll(i.InstSubDir(name))
}

func (i *CielDir) InstExists(name string) bool {
	if path.Clean(i.InstSubDir(name)) == path.Clean(i.InstDir()) {
		return false
	}
	if stat, err := os.Stat(i.InstSubDir(name)); err != nil || !stat.IsDir() {
		return false
	}
	return true
}

func (i *CielDir) InstFileSystem(name string) overlayfs.Instance {
	layersDir := path.Join(i.InstSubDir(name), LayerDirName)
	var layers = []string{i.DistDir()}
	for _, dir := range Layers {
		layers = append(layers, path.Join(layersDir, dir))
	}
	return overlayfs.Instance{
		MountPoint: "./" + name,
		Layers:     layers,
	}
}

func (i *CielDir) InstMount(name string) error {
	ofs := i.InstFileSystem(name)
	if !utils.Lock(i.InstLockFile(name)) {
		if i.InstMounted(name) {
			return ErrLock
		}
		utils.Unlock(i.InstLockFile(name))
	}
	if err := ofs.Mount(false); err != nil {
		utils.Unlock(i.InstLockFile(name))
		return err
	}
	return nil
}

func (i *CielDir) InstMounted(name string) bool {
	a, err := ioutil.ReadFile("/proc/self/mountinfo")
	s := string(a)
	list := strings.Split(s, "\n")
	match, _ := filepath.Abs(i.InstMountPoint(name))
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

func (i *CielDir) InstUnmount(name string) error {
	i.InstUnlockContainer(name)
	ofs := i.InstFileSystem(name)
	if err := ofs.Unmount(); err != nil {
		return err
	}
	os.Remove(ofs.MountPoint)
	utils.Unlock(i.InstLockFile(name))
	return nil
}

func (i *CielDir) InstRun(ctx context.Context, name string, boot bool, network bool, containerArgs []string, args ...string) (int, error) {
	defer RecoverTerminalAttr()
	machineId := name + "_" + utils.RandomString(5)

	if !i.InstRefractoryPeriodEnter(name) {
		return -1, errors.New("another instance is in refractory period")
	}

	if boot && nspawn.IsBootable(name) {
		toBoot := false
		if oldMachineId := i.InstMachineId(name); oldMachineId == "" {
			i.InstBootedEnter(name)
			i.InstSetMachineId(name, machineId)
			toBoot = true
		} else if !nspawn.MachineRunning(nspawn.MachineStatus(ctx, oldMachineId)) {
			i.InstUnlockContainer(name)
			i.InstBootedEnter(name)
			i.InstSetMachineId(name, machineId)
			toBoot = true
		} else {
			machineId = oldMachineId
		}
		i.InstRefractoryPeriodLeave(name)
		if toBoot {
			if network {
				containerArgs = append([]string{"--network-zone=ciel"}, containerArgs...)
			}
			_, err := nspawn.SystemdNspawn(ctx, i.InstMountPoint(name), true, machineId, containerArgs...)

			// NOTE: This will be blocked until the container completely booted up.
			if _, ok := err.(nspawn.ErrCancelled); ok {
				i.InstUnsetMachineId(name)
				i.InstBootedLeave(name)
				return -1, err
			}
		}
		return nspawn.SystemdRun(ctx, machineId, args...)
	}

	i.InstSetMachineId(name, machineId)

	defer i.InstUnsetMachineId(name)
	defer i.InstRefractoryPeriodLeave(name)

	return nspawn.SystemdNspawn(ctx, i.InstMountPoint(name), false, machineId, args...)
}

func (i *CielDir) InstStop(ctx context.Context, name string) error {
	if i.InstMachineId(name) == "" {
		return errors.New("no machineId")
	}
	var err error
	if i.InstBooted(name) {
		err = nspawn.MachinectlPoweroff(ctx, i.InstMachineId(name))
		if err == nil {
			i.InstBootedLeave(name)
			i.InstUnsetMachineId(name)
		}
	} else {
		err = nspawn.MachinectlTerminate(ctx, i.InstMachineId(name))
		if err == nil {
			i.InstUnsetMachineId(name)
			i.InstRefractoryPeriodLeave(name)
		}
	}
	return err
}

const cRST = "\x1b[0m"
const cGRE = "\x1b[92m"
const cGRY = "\x1b[37m"
const cRED = "\x1b[91m"
const cCYA = "\x1b[96m"

func (i *CielDir) InstFileSystemStat(name string) string {
	if i.InstMounted(name) {
		return cGRE + "mounted" + cRST
	}
	return cRST + "free" + cRST
}

func (i *CielDir) InstContainerRunningStat(name string) (status, boot string) {
	if utils.Locked(i.InstRefractoryFile(name)) {
		return cRED + "locked" + cRST, cGRY + "unknown" + cRST
	}
	if i.InstMachineId(name) != "" {
		if i.InstBooted(name) {
			status := nspawn.MachineStatus(context.Background(), i.InstMachineId(name))
			if !nspawn.MachineRunning(status) {
				if nspawn.MachineDead(status) {
					i.InstBootedLeave(name)
					i.InstUnsetMachineId(name)
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

func (i *CielDir) InstLockFile(name string) string {
	return path.Join(i.InstSubDir(name), LockFileName)
}

func (i *CielDir) InstBootedFile(name string) string {
	return path.Join(i.InstSubDir(name), BootedFileName)
}

func (i *CielDir) InstBootedEnter(name string) bool {
	return utils.Lock(i.InstBootedFile(name))
}

func (i *CielDir) InstBooted(name string) bool {
	return utils.Locked(i.InstBootedFile(name))
}

func (i *CielDir) InstBootedLeave(name string) {
	utils.Unlock(i.InstBootedFile(name))
}

func (i *CielDir) InstRefractoryFile(name string) string {
	return path.Join(i.InstSubDir(name), RefractoryFileName)
}

func (i *CielDir) InstRefractoryPeriodEnter(name string) bool {
	return utils.Lock(i.InstRefractoryFile(name))
}

func (i *CielDir) InstRefractoryPeriodLeave(name string) {
	utils.Unlock(i.InstRefractoryFile(name))
}

func (i *CielDir) InstMachineIdFile(name string) string {
	return path.Join(i.InstSubDir(name), MachineIdFileName)
}

func (i *CielDir) InstSubDir(name string) string {
	return path.Join(i.InstDir(), name)
}

func (i *CielDir) InstMachineId(name string) string {
	b, _ := ioutil.ReadFile(i.InstMachineIdFile(name))
	return strings.TrimSpace(string(b))
}

func (i *CielDir) InstSetMachineId(name, machineId string) error {
	return ioutil.WriteFile(i.InstMachineIdFile(name), []byte(machineId), 0644)
}

func (i *CielDir) InstUnsetMachineId(name string) error {
	return os.Remove(i.InstMachineIdFile(name))
}

func (i *CielDir) InstMountPoint(name string) string {
	return path.Join(i.BasePath, name)
}

func (i *CielDir) InstUnlockFileSystem(name string) {
	i.InstUnmount(name)
	os.Remove(i.InstLockFile(name))
	os.Remove(i.InstMountPoint(name))
}

func (i *CielDir) InstUnlockContainer(name string) {
	i.InstStop(context.Background(), name)
	os.Remove(i.InstBootedFile(name))
	os.Remove(i.InstMachineIdFile(name))
	os.Remove(i.InstRefractoryFile(name))

}

func (i *CielDir) GetAll() []string {
	dir, err := os.Open(i.InstDir())
	if err != nil {
		log.Panic(err)
	}
	subDirs, err := dir.Readdir(0)
	if err != nil {
		log.Panic(err)
	}
	var subDirNames []string
	for _, subDirs := range subDirs {
		if subDirs.IsDir() {
			subDirNames = append(subDirNames, subDirs.Name())
		}
	}
	return subDirNames
}

func (i *CielDir) MountAll() {
	instList := i.GetAll()
	for _, inst := range instList {
		err := i.InstMount(inst)
		if err != nil {
			log.Println(inst+":", err)
		} else {
			log.Println(inst+":", "done")
		}
	}
}

func (i *CielDir) UnmountAll() {
	instList := i.GetAll()
	for _, inst := range instList {
		err := i.InstUnmount(inst)
		if err != nil {
			log.Println(inst+":", err)
		} else {
			log.Println(inst+":", "done")
		}
	}
}
