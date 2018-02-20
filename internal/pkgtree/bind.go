package pkgtree

import (
	"os"
	"syscall"

	"ciel/display"
)

func (t *Tree) Mount(mountPoint string) {
	if _, err := os.Stat(t.BasePath); os.IsNotExist(err) {
		return
	}
	os.MkdirAll(mountPoint, 0755)
	syscall.Mount(t.BasePath, mountPoint, "", syscall.MS_BIND, "")
}

func (t *Tree) Unmount(mountPoint string) {
	if _, err := os.Stat(mountPoint); os.IsNotExist(err) {
		return
	}
	d.ITEM("unmount tree")
	err := syscall.Unmount(mountPoint, 0)
	d.ERR(err)
}
