package packaging

import (
	"os"
	"path"

	"ciel/display"
	"ciel/internal/abstract"
)

const (
	AB3Path  = "/usr/bin/autobuild"
	ACBSPath = "/usr/bin/acbs-build"
)

func DetectToolChain(i abstract.Instance) {
	root := i.MountPoint()
	d.ITEM("detect autobuild3")
	exists(root, AB3Path)
	d.ITEM("detect acbs")
	exists(root, ACBSPath)
}

func exists(root, target string) bool {
	_, err := os.Stat(path.Join(root, target))
	if os.IsNotExist(err) {
		d.FAILED()
		return false
	} else if err == nil {
		d.OK()
		return true
	} else {
		d.FAILED_BECAUSE(err.Error())
		return false
	}
}
