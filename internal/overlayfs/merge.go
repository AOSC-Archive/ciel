package overlayfs

import (
	"errors"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

func (i *Instance) file2any(rpath string, src, dst int) error {
	// file dir
	// | \
	// |  \
	// file dir

	upath := filepath.Join(i.Layers[src], rpath)
	lpath := filepath.Join(i.Layers[dst], rpath)
	os.RemoveAll(lpath)
	utype, _ := overlayTypeByLstat(upath)
	if !(dst == 0 && utype == overlayTypeWhiteout) {
		// if it is NOT to put a whiteout onto bottom
		return os.Rename(upath, lpath)
	}
	return nil
}

func (i *Instance) dir2dir(rpath string, src, dst int) error {
	// file dir
	//       |
	//       |
	// file dir

	// copy attributes, and continue.
	upath := filepath.Join(i.Layers[src], rpath)
	lpath := filepath.Join(i.Layers[dst], rpath)
	return copyAttributes(upath, lpath)
}

func (i *Instance) dir2file(rpath string, src, dst int) error {
	// file dir
	//     /
	//    /
	// file dir

	// the upper layer is a directory,
	// the lower layer is a whiteout or a normal file, which can be a cover,
	// that removing them may let the content in lower layers appear.

	// if the lower layer is at the bottom
	// or lower layers under the lower layer have another cover,
	// we can merge the upper one safely.
	upath := filepath.Join(i.Layers[src], rpath)
	lpath := filepath.Join(i.Layers[dst], rpath)
	nextfilelayer, havedir := i.nextLayerHasFile(rpath, dst)
	if !havedir {
		os.Remove(lpath)
		if err := os.Rename(upath, lpath); err != nil {
			return err
		}
		return filepath.SkipDir
	}

	// 1). "open" the directory
	os.Mkdir(lpath, 0755)
	if err := copyAttributes(upath, lpath); err != nil {
		return err
	}
	// 2). "cover" all sub-files in the directory
	for filename := range i.readDirInRange(rpath, dst-1, nextfilelayer+1) {
		createWhiteout(filepath.Join(lpath, filename))
	}
	return nil
}

func (i *Instance) Merge() error {
	return i.merge("/", len(i.Layers)-1, 0)
}

// MergeFile is the method to merge a file or directory from an upper layer
// to a lower layer.
func (i *Instance) merge(path string, src, dst int) error {
	path = filepath.Clean(path)
	uroot, lroot := i.Layers[src], i.Layers[dst]
	walkBase := filepath.Join(uroot, path)
	os.MkdirAll(filepath.Dir(filepath.Join(lroot, path)), 755)
	err := filepath.Walk(walkBase, func(upath string, info os.FileInfo, err error) error {
		log.Println(upath)
		rel, _ := filepath.Rel(uroot, upath)
		lpath := filepath.Join(lroot, rel)

		utp, err := overlayTypeByInfo(info, err)
		if err != nil {
			return err
		}
		ltp, err := overlayTypeByLstat(lpath)
		if err != nil {
			return err
		}

		switch utp {
		case overlayTypeAir:
			return nil

		case overlayTypeFile, overlayTypeWhiteout:
			return i.file2any(rel, src, dst)

		case overlayTypeDir:
			switch ltp {
			case overlayTypeAir:
				if err := os.Rename(upath, lpath); err != nil {
					return err
				}
				return filepath.SkipDir

			case overlayTypeDir:
				return i.dir2dir(rel, src, dst)

			case overlayTypeFile, overlayTypeWhiteout:
				return i.dir2file(rel, src, dst)
			}
		}
		return errors.New("unexpected type")
		// end of walk-function
	})
	return err
}

type overlayType string

const (
	overlayTypeInvalid overlayType = ""

	overlayTypeAir      = "-"
	overlayTypeWhiteout = "x"
	overlayTypeFile     = "f"
	overlayTypeDir      = "d"
)

func copyAttributes(src, dst string) error {
	args := []string{
		"--preserve=all",
		"--attributes-only",
		"--no-target-directory",
		"--recursive",
		"--no-clobber",
		src,
		dst,
	}
	cmd := exec.Command("/bin/cp", args...)
	a, err := cmd.CombinedOutput()
	if err != nil {
		log.Println("copy attributes: ", strings.TrimSpace(string(a)))
		// FIXME: failed?
	}
	return nil
}

func createWhiteout(path string) error {
	return syscall.Mknod(path, 0000, 0x0000)
}

func overlayTypeByLstat(path string) (overlayType, error) {
	return overlayTypeByInfo(os.Lstat(path))
}

func overlayTypeByInfo(info os.FileInfo, err error) (overlayType, error) {
	if os.IsNotExist(err) {
		return overlayTypeAir, nil
	} else if err != nil {
		return overlayTypeInvalid, err
	}
	if info.IsDir() {
		return overlayTypeDir, nil
	}
	if isWhiteout(info) {
		return overlayTypeWhiteout, nil
	}
	return overlayTypeFile, nil
}

func isWhiteout(fi os.FileInfo) bool {
	const mask = os.ModeDevice | os.ModeCharDevice
	if fi.Mode()&mask != mask {
		return false
	}
	return fi.Sys().(*syscall.Stat_t).Rdev == 0
}

func (i *Instance) nextLayerHasFile(relpath string, cur int) (layer int, hasdir bool) {
	layer = -1
	hasdir = false
	if cur != 0 {
		for index := cur - 1; index >= 0; index-- {
			iroot := i.Layers[index]
			ipath := filepath.Join(iroot, relpath)
			itp, _ := overlayTypeByLstat(ipath)
			switch itp {
			case overlayTypeFile, overlayTypeWhiteout:
				layer = index
				return
			case overlayTypeDir:
				hasdir = true
			}
		}
	}
	return
}

func (i *Instance) readDirInRange(relpath string, ubound, lbound int) map[string]bool {
	filelist := make(map[string]bool)
	for index := lbound; index <= ubound; index++ {
		iroot := i.Layers[index]
		ipath := filepath.Join(iroot, relpath)
		iinfo, err := os.Lstat(ipath)
		if os.IsNotExist(err) {
			continue
		}
		idir, err := os.Open(ipath)
		if err != nil {
			continue
		}
		iinfos, err := idir.Readdir(0) // 0: check all sub-files
		if err != nil {
			continue
		}
		for _, iiinfo := range iinfos {
			iitp, _ := overlayTypeByInfo(iiinfo, nil)
			if iitp == overlayTypeWhiteout {
				delete(filelist, iinfo.Name())
			} else {
				filelist[iinfo.Name()] = true
			}
		}
		idir.Close()
	}
	return filelist
}
