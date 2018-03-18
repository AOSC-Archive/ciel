package pkgtree

import (
	"os"
	"syscall"

	"ciel/display"
	"path"
)

const (
	TreePath = "/tree"
)

func (t *Tree) Mount(mountPoint string) {
	if _, err := os.Stat(t.BasePath); os.IsNotExist(err) {
		return
	}
	treeMountPoint := path.Join(mountPoint, TreePath)
	os.MkdirAll(treeMountPoint, 0755)
	syscall.Mount(t.BasePath, treeMountPoint, "", syscall.MS_BIND, "")
}

func (t *Tree) Unmount(mountPoint string) {
	treeMountPoint := path.Join(mountPoint, TreePath)
	if _, err := os.Stat(treeMountPoint); os.IsNotExist(err) {
		return
	}
	d.ITEM("unmount tree")
	err := syscall.Unmount(treeMountPoint, 0)
	d.ERR(err)
	if err != nil {
		return
	}
	d.ITEM("remove tree mount point")
	err = os.Remove(treeMountPoint)
	d.ERR(err)
}
