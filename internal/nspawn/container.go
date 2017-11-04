package nspawn

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"
	"time"
)

var BootableFiles = []string{
	"/usr/lib/systemd/systemd",
	"/lib/systemd/systemd",
	"/sbin/init",
}

const PoweroffTimeout = 5 * time.Second

var (
	ErrCancelled = errors.New("cancelled")
)

func IsBootable(p string) bool {
	for _, file := range BootableFiles {
		_, err := os.Stat(path.Join(p, file))
		if err == nil {
			return true
		}
	}
	return false
}

func SystemdNspawn(ctx context.Context, directory string, boot bool, machineId string, args ...string) (int, error) {
	a := []string{
		"--quiet",
		"-D", directory,
	}
	if boot {
		a = append(a, "--boot")
	}
	if machineId != "" {
		a = append(a, "-M", machineId)
	}
	a = append(a, args...)
	cmd := exec.CommandContext(ctx, "systemd-nspawn", a...)
	if !boot {
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin
	}

	var err error
	if boot {
		waitCtx, cancelFunc := context.WithCancel(context.Background())
		go func() {
			cmd.Run()
			cancelFunc()
		}()
		cancelled := waitUntilRunningOrDegraded(waitCtx, machineId)
		if cancelled {
			return -1, ErrCancelled
		}
	} else {
		err = cmd.Run()
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.Sys().(syscall.WaitStatus).ExitStatus(), nil
		}
		if err != nil {
			return -1, err
		}
	}

	return 0, nil
}

func SystemdRun(ctx context.Context, machineId string, args ...string) (int, error) {
	a := []string{
		"--quiet",
		"--wait",
		"--pty",
		"-M", machineId,
	}
	a = append(a, args...)
	cmd := exec.CommandContext(ctx, "systemd-run", a...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	err := cmd.Run()
	if exitErr, ok := err.(*exec.ExitError); ok {
		return exitErr.Sys().(syscall.WaitStatus).ExitStatus(), nil
	} else {
		return -1, err
	}
}

func MachinectlTerminate(ctx context.Context, machineId string) error {
	err := machinectlTerminate(ctx, machineId)
	waitUntilShutdown(ctx, machineId)
	return err
}

func MachinectlPoweroff(ctx context.Context, machineId string) error {
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
	waitCtx, _ := context.WithTimeout(ctx, PoweroffTimeout)
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

func waitUntilRunningOrDegraded(ctx context.Context, machindId string) (cancelled bool) {
	for {
		switch {
		case MachineRunning(ctx, machindId):
			return false
		default:
			if ctx != nil {
				select {
				case <-ctx.Done():
					return true
				default:
				}
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func waitUntilShutdown(ctx context.Context, machindId string) (cancelled bool) {
	for {
		switch {
		case MachineDead(ctx, machindId):
			return false
		default:
			if ctx != nil {
				select {
				case <-ctx.Done():
					return true
				default:
				}
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func MachineStatus(ctx context.Context, machindId string) string {
	a := []string{
		"is-system-running",
		"-M", machindId,
	}
	cmd := exec.CommandContext(ctx, "systemctl", a...)
	cmd.Env = append(os.Environ(), "LANG=C")
	output, _ := cmd.CombinedOutput()
	return strings.TrimSpace(string(output))
}

func MachineRunning(ctx context.Context, machindId string) bool {
	switch MachineStatus(ctx, machindId) {
	case "running":
		return true
	case "degraded":
		return true
	default:
		return false
	}
}

func MachineDead(ctx context.Context, machindId string) bool {
	switch MachineStatus(ctx, machindId) {
	case "Failed to connect to bus: Host is down":
		return true
	default:
		return false
	}
}
