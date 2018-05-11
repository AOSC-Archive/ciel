package nspawn

import (
	"context"
	"time"
)

func waitUntilRunningOrDegraded(ctx context.Context, machineId string) (cancelled bool) {
	for {
		switch {
		case MachineRunning(MachineStatus(context.TODO(), machineId)):
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

func waitUntilShutdown(ctx context.Context, machineId string) (cancelled bool) {
	for {
		switch {
		case MachineDead(MachineStatus(ctx, machineId)):
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
