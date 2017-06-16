package main

import (
	"os"
	"path/filepath"
	"strings"
)

var cacheLinks map[string]bool

func evalSymlinks(root string, path string, noleaf bool) string {
	path = filepath.Clean(path)
	var pos int
	var prefix string
	for {
		if pos == len(path) {
			break
		}
		if subpos := strings.Index(path[pos:], "/"); subpos == -1 {
			if noleaf {
				break
			} else {
				pos = len(path)
				prefix = path
			}
		} else {
			pos += subpos + 1
			if pos == 1 {
				prefix = path[:pos]
			} else {
				prefix = path[:pos-1]
			}
		}
		isLink, cached := cacheLinks[prefix]
		if cached {
			if !isLink {
				continue
			} else {
				panic("loop in symlink")
			}
		} else {
			fi, _ := os.Lstat(filepath.Join(root, prefix))
			if fi == nil || fi.Mode()&os.ModeSymlink == 0 {
				cacheLinks[prefix] = false
				continue
			} else {
				cacheLinks[prefix] = true
			}
			target, _ := os.Readlink(filepath.Join(root, prefix))
			if !filepath.IsAbs(target) {
				target = filepath.Join(filepath.Dir(prefix), target)
			}
			if cleanTarget := evalSymlinks(root, target, noleaf); cleanTarget != target {
				target = cleanTarget
			}
			path = filepath.Join(target, path[pos:])
			pos = len(target)
		}
	}
	return filepath.Clean(path)
}

func evalSymlinksClean() {
	cacheLinks = make(map[string]bool)
}
