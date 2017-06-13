package main

import (
	"ciel-driver"
	"log"
	"math/rand"
	"time"
)

func main() {
	c := ciel.New("buildkit", "/root/ciel/test/container-layers")
	defer func() {
		c.Unmount()
	}()
	defer func() {
		c.Shutdown()
	}()

	var exitCode int
	println("[ciel] apt update")
	exitCode = c.Command("apt update -y")
	if exitCode != 0 {
		log.Panicln("exit code:", exitCode)
	}
	println("\n[ciel] apt full-upgrade")
	exitCode = c.Command("apt full-upgrade -y")
	if exitCode != 0 {
		log.Panicln("exit code:", exitCode)
	}
	println("\n[ciel] apt install systemd")
	exitCode = c.Command("apt install -y systemd")
	if exitCode != 0 {
		log.Panicln("exit code:", exitCode)
	}
	println("\n[ciel] apt install {base}")
	exitCode = c.Command("apt install -y admin-base core-base editor-base python-base" +
		" network-base systemd-base web-base util-base devel-base debug-base autobuild3 git")
	if exitCode != 0 {
		log.Panicln("exit code:", exitCode)
	}
	c.Shutdown()
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
