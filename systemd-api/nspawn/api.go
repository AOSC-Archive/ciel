package nspawn

import (
	"bytes"
	"context"
	"errors"
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

func SystemdNspawnRun(ctx context.Context, directory string, machineId string, args ...string) (int, error) {
	a := []string{
		"--quiet",
		"-D", directory,
	}
	if machineId != "" {
		a = append(a, "-M", machineId)
	}
	a = append(a, args...)
	cmd := exec.CommandContext(ctx, "systemd-nspawn", a...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	err := cmd.Run()
	if exitErr, ok := err.(*exec.ExitError); ok {
		return exitErr.Sys().(syscall.WaitStatus).ExitStatus(), nil
	}
	if err != nil {
		return -1, err
	}

	return 0, nil
}
func SystemdNspawnBoot(ctx context.Context, directory string, machineId string, args ...string) (int, error) {
	a := []string{
		"--quiet",
		"--boot",
		"-D", directory,
	}
	if machineId != "" {
		a = append(a, "-M", machineId)
	}
	a = append(a, args...)
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
		return -1, err
	}
	return 0, nil
}

func SystemdRun(ctx context.Context, machineId string, args ...string) (int, error) {
	a := []string{
		"--quiet",
		"--wait",
		"--pty",
		"--uid=root", // for login sessions
		"-M", machineId,
	}
	a = append(a, args...)
	cmd := exec.CommandContext(ctx, "systemd-run", a...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	err := cmd.Run()

	defer func() {
		// shutting down...
		if !MachineRunning(MachineStatus(ctx, machineId)) {
			waitUntilShutdown(ctx, machineId)
		}
	}()

	if exitErr, ok := err.(*exec.ExitError); ok {
		return exitErr.Sys().(syscall.WaitStatus).ExitStatus(), nil
	}
	if err != nil {
		return -1, err
	}
	return 0, nil
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
