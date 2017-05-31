package main

import (
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func redirectStdIO(cmd *exec.Cmd) {
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
}

func pipeIO(r io.Reader, w io.Writer) {
	go func() {
		for {
			if _, err := io.Copy(w, r); err != nil {
				return
			}
		}
	}()
}

func findBuildLog(path string) string {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return ""
	}
	s := string(b)
	const prefix = " Build log: "
	const suffix = "\n"
	prefixindex := strings.LastIndex(s, prefix)
	if prefixindex == -1 {
		return ""
	}
	s = s[prefixindex+len(prefix):]
	suffixindex := strings.Index(s, suffix)
	if suffixindex == -1 {
		return ""
	}
	s = s[:suffixindex]
	return s
}
