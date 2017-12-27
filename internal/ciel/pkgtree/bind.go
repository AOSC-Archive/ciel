package pkgtree

import (
	"os"
	"syscall"

	"ciel/internal/display"
)

func (t *Tree) Mount(mountPoint string) {
	os.MkdirAll(mountPoint, 0755)
	syscall.Mount(t.BasePath, mountPoint, "", syscall.MS_BIND, "")
}

func (t *Tree) Unmount(mountPoint string) {
	d.ITEM("unmount tree")
	syscall.Unmount(mountPoint, 0)
	d.OK()
}
