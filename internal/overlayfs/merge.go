package overlayfs

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

func removeIfExist(path string) error {
	err := os.RemoveAll(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func override(upPath, lowPath string) error {
	if err := removeIfExist(lowPath); err != nil {
		return err
	}
	return os.Rename(upPath, lowPath)
}

func removeBoth(upPath, lowPath string) error {
	if err := removeIfExist(lowPath); err != nil {
		return err
	}
	return os.Remove(upPath)
}

// MergeFile is the method to merge a file or directory from an upper layer
// to a lower layer.
func (i *Instance) Merge() error {
	upRoot, lowRoot := i.Layers[len(i.Layers)-1], i.Layers[0]
	err := filepath.Walk(upRoot, func(upPath string, info os.FileInfo, err error) error {
		relPath, _ := filepath.Rel(upRoot, upPath)
		lowPath := filepath.Join(lowRoot, relPath)

		upType, err := overlayTypeByInfo(info, err)
		if err != nil {
			return err
		}
		lowType, err := overlayTypeByLstat(lowPath)
		if err != nil {
			return err
		}

		//    n  f  w  d
		// n  -  o  r  r (skip sub-directories)
		// f  -  o  r  o
		// w  -  o  r  o
		// d  -  o  r  c

		// o = override
		// r = remove both
		// m = move
		// c = copy attributes
		// s = skip sub-directories

		switch upType {
		case overlayTypeNothing:
			return nil

		case overlayTypeFile:
			return override(upPath, lowPath)

		case overlayTypeWhiteout:
			return removeBoth(upPath, lowPath)

		case overlayTypeDir:
			switch lowType {
			case overlayTypeNothing:
				if err := override(upPath, lowPath); err != nil {
					return err
				}
				return filepath.SkipDir

			case overlayTypeFile:
				if err := override(upPath, lowPath); err != nil {
					return err
				}
				return filepath.SkipDir

			case overlayTypeWhiteout:
				// strange case. a whiteout file in the lowest layer?
				if err := override(upPath, lowPath); err != nil {
					return err
				}
				return filepath.SkipDir

			case overlayTypeDir:
				err := copyAttributes(upPath, lowPath)
				// remove empty directory
				if err := os.Remove(upPath); err == nil {
					return filepath.SkipDir
				}
				return err
			}
		}
		panic("unexpected type")
	})
	// end of walk-function
	if err == nil {
		list, err := ioutil.ReadDir(upRoot)
		if err != nil {
			return err
		}
		if len(list) != 0 {
			for _, info := range list {
				os.RemoveAll(info.Name())
			}
		}
	}
	return err
}

type overlayType int

const (
	overlayTypeNothing overlayType = iota
	overlayTypeFile
	overlayTypeWhiteout
	overlayTypeDir
)

func copyAttributes(src, dst string) error {
	// TODO: copying attributes, NOT recursively
	args := []string{
		"--preserve=all",
		"--attributes-only",
		"--no-target-directory",
		"--recursive", // BUG: cp cannot do this not recursively
		"--no-clobber",
		src,
		dst,
	}
	exec.Command("/bin/cp", args...).Run()
	return nil
}

func overlayTypeByLstat(path string) (overlayType, error) {
	return overlayTypeByInfo(os.Lstat(path))
}

func overlayTypeByInfo(info os.FileInfo, err error) (overlayType, error) {
	if os.IsNotExist(err) {
		return overlayTypeNothing, nil
	} else if err != nil {
		return overlayTypeNothing, err
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
