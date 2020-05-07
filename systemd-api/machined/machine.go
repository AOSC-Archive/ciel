package machined

import (
	"github.com/godbus/dbus/v5"
)

type Machine struct {
	Obj dbus.BusObject
}

const MachineInterface = "org.freedesktop.machine1.Machine"

func (m Machine) Leader() (uint32, error) {
	v, err := m.GetProperty(".Leader")
	if err != nil {
		return 0, err
	}
	return v.(uint32), err
}

func (m Machine) GetProperty(name string) (value interface{}, err error) {
	v, err := m.Obj.GetProperty(MachineInterface + name)
	if err != nil {
		return nil, err
	}
	return v.Value(), err
}
