package packaging

import (
	"os"
	"path"

	"github.com/AOSC-Dev/ciel/internal/abstract"

	d "github.com/AOSC-Dev/ciel/display"
)

const (
	AB3Path  = "/usr/bin/autobuild"
	ACBSPath = "/usr/bin/acbs-build"
)

type ToolChain struct {
	AB   bool
	ACBS bool
}

func DetectToolChain(global bool, i abstract.Instance, c abstract.Container) *ToolChain {
	var root string
	if global {
		root = c.DistDir()
	} else {
		root = i.MountPoint()
	}
	tc := &ToolChain{}
	d.ITEM("detect autobuild3")
	tc.AB = exists(root, AB3Path)
	d.ITEM("detect acbs")
	tc.ACBS = exists(root, ACBSPath)
	return tc
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
