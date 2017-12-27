package nspawn

import (
	"context"
	"os"
	"os/exec"
	"path"
	"strings"
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

func MachineRunning(status string) bool {
	switch status {
	case "running":
		return true
	case "degraded":
		return true
	default:
		return false
	}
}

func MachineDead(status string) bool {
	switch status {
	case "Failed to connect to bus: Host is down":
		return true
	default:
		return false
	}
}
