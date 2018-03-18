package nspawn

import (
	"bytes"
	"context"
	"errors"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

var BootableFiles = []string{
	"/usr/lib/systemd/systemd",
	"/lib/systemd/systemd",
	"/sbin/init",
}

const PowerOffTimeout = 30 * time.Second

type ErrCancelled struct {
	reason string
}

func (e ErrCancelled) Error() string {
	return "cancelled: " + e.reason
}

func SystemdNspawnRun(ctx context.Context, machineId string, dir string, ctnInfo *ContainerInfo, runInfo *RunInfo) (int, error) {
	a := nspawnArgs(machineId, dir, ctnInfo, runInfo)
	cmd := exec.CommandContext(ctx, "systemd-nspawn", a...)
	setCmdStdDev(cmd, runInfo.StdDev)

	err := cmd.Run()
	return unpackExecErr(err)
}
func SystemdNspawnBoot(ctx context.Context, machineId string, dir string, ctnInfo *ContainerInfo) error {
	a := nspawnArgs(machineId, dir, ctnInfo, nil)
	cmd := exec.CommandContext(ctx, "systemd-nspawn", a...)

	var err error
	waitCtx, cancelFunc := context.WithCancel(context.Background())
	go func() {
		errBuf := &bytes.Buffer{}
		cmd.Stderr = errBuf
		cmd.Run()
		output := errBuf.String()
		err = ErrCancelled{reason: string(output)}
		cancelFunc()
	}()
	cancelled := waitUntilRunningOrDegraded(waitCtx, machineId)
	if cancelled {
		return err
	}
	return nil
}

func SystemdRun(ctx context.Context, machineId string, runInfo *RunInfo) (int, error) {
	a := runArgs(machineId, runInfo)
	cmd := exec.CommandContext(ctx, "systemd-run", a...)
	setCmdStdDev(cmd, runInfo.StdDev)

	err := cmd.Run()
	defer func() {
		// shutting down...
		if !MachineRunning(MachineStatus(ctx, machineId)) {
			waitUntilShutdown(ctx, machineId)
		}
	}()
	return unpackExecErr(err)
}

func MachinectlShell(ctx context.Context, machineId string, runInfo *RunInfo) (int, error) {
	a := msArgs(machineId, runInfo)
	cmd := exec.CommandContext(ctx, "machinectl", a...)
	setCmdStdDev(cmd, runInfo.StdDev)

	err := cmd.Run()
	defer func() {
		// shutting down...
		if !MachineRunning(MachineStatus(ctx, machineId)) {
			waitUntilShutdown(ctx, machineId)
		}
	}()
	return unpackExecErr(err)
}

func MachinectlTerminate(ctx context.Context, machineId string) error {
	err := machinectlTerminate(ctx, machineId)
	waitUntilShutdown(ctx, machineId)
	return err
}

func MachinectlPowerOff(ctx context.Context, machineId string) error {
	a := []string{
		"poweroff",
		"--quiet",
		machineId,
	}
	cmd := exec.CommandContext(ctx, "machinectl", a...)
	output, err := cmd.CombinedOutput()
	if _, ok := err.(*exec.ExitError); ok {
		return errors.New(strings.TrimSpace(string(output)))
	} else if err != nil {
		return err
	}
	waitCtx, _ := context.WithTimeout(ctx, PowerOffTimeout)
	if waitUntilShutdown(waitCtx, machineId) { // cancelled
		machinectlTerminate(context.Background(), machineId)
	}
	return nil
}

func machinectlTerminate(ctx context.Context, machineId string) error {
	a := []string{
		"terminate",
		"--quiet",
		machineId,
	}
	cmd := exec.CommandContext(ctx, "machinectl", a...)
	output, err := cmd.CombinedOutput()
	if _, ok := err.(*exec.ExitError); ok {
		return errors.New(strings.TrimSpace(string(output)))
	} else {
		return err
	}
}

func nspawnArgs(machineId string, dir string, ctnInfo *ContainerInfo, runInfo *RunInfo) []string {
	if machineId == "" {
		log.Panicln("no machineId specified")
	}

	a := []string{
		"--quiet",
		"-D", dir,
		"-M", machineId,
	}

	for _, v := range ctnInfo.Properties {
		a = append(a, "--property="+v)
	}

	if ctnInfo.Init {
		a = append(a, "--boot")
	}
	if ctnInfo.Network != nil {
		netInfo := ctnInfo.Network
		if netInfo.Zone != "" {
			a = append(a, "--network-zone="+netInfo.Zone)
		}
	}

	a = append(a, "--")

	if ctnInfo.Init {
		a = append(a, ctnInfo.InitArgs...)
	} else {
		a = append(a, runInfo.App)
		a = append(a, runInfo.Args...)
	}

	return a
}

func runArgs(machineId string, runInfo *RunInfo) []string {
	a := []string{
		"--quiet",
		"--wait",
		"--pty",
		"--uid=root", // for login sessions
		"--send-sighup",
		"-M", machineId,
	}
	a = append(a, "--")
	a = append(a, runInfo.App)
	a = append(a, runInfo.Args...)
	return a
}

func msArgs(machineId string, runInfo *RunInfo) []string {
	a := []string{
		"shell",
		"--quiet",
		machineId,
	}
	a = append(a, runInfo.App)
	a = append(a, runInfo.Args...)
	return a
}

func setCmdStdDev(cmd *exec.Cmd, stdDev *StdDevInfo) {
	if stdDev == nil {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stdin = stdDev.Stdin
		cmd.Stdout = stdDev.Stdout
		cmd.Stderr = stdDev.Stderr
	}
}

func unpackExecErr(err error) (int, error) {
	if exitErr, ok := err.(*exec.ExitError); ok {
		return exitErr.Sys().(syscall.WaitStatus).ExitStatus(), nil
	}
	if err != nil {
		return -1, err
	}
	return 0, nil
}
