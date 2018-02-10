package pkgtree

import (
	"path"

	"ciel/internal/abstract"
)

const (
	TreePath = "/tree"
)

func (t *Tree) MountHandler(i abstract.Instance, mount bool) {
	treeMountPoint := path.Join(i.MountPoint(), TreePath)
	if mount {
		t.Mount(treeMountPoint)
	} else {
		t.Unmount(treeMountPoint)
	}
}
