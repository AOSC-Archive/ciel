package main

import (
	"ciel-driver"
	"log"
	"math/rand"
	"time"
)

func main() {
	c := ciel.New("buildkit", "/root/ciel/test/container-layers")
	defer c.Unmount()
	defer c.Shutdown()

	if err := updateStub(c); err != nil {
		log.Panicln("clean stub error:", err)
	}

	if ec := c.Command("apt install -y systemd"); ec != 0 {
		log.Panicln("apt install systemd exit code:", ec)
	}
	if ec := c.Command("apt install -y admin-base core-base editor-base python-base" +
		" network-base systemd-base web-base util-base devel-base debug-base autobuild3 git"); ec != 0 {
		log.Panicln("apt install {base} exit code:", ec)
	}
	if err := cleanStub(c); err != nil {
		log.Panicln("clean stub error:", err)
	}
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
