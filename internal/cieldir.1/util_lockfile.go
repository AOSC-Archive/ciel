package cieldir

import (
	"os"
)

func lock(path string) bool {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0)
	if err != nil {
		return false
	}
	f.Close()
	return true
}

func unlock(path string) bool {
	return os.Remove(path) == nil
}

func locked(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
