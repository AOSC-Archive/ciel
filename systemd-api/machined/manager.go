package machined

import (
	"github.com/godbus/dbus"
)

const ManagerPath = "/org/freedesktop/machine1"

type Manager struct {
	Obj dbus.BusObject
}

const ManagerInterface = "org.freedesktop.machine1.Manager"

func (m Manager) GetMachine(name string) (machine *Machine, err error) {
	result := m.Obj.Call(ManagerInterface+".GetMachine", 0, name)
	if result.Err != nil {
		return nil, result.Err
	}
	return &Machine{Object(result.Body[0].(dbus.ObjectPath))}, nil
}

func NewManager() *Manager {
	return &Manager{Object(ManagerPath)}
}
