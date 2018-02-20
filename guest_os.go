package main

import (
	"context"
	"flag"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"ciel/display"
	"ciel/internal/ciel"
)

const (
	LatestTarballURL = "https://repo.aosc.io/aosc-os/os-amd64/buildkit/aosc-os_buildkit_20180128_amd64.tar.xz"
	DownloadTarball  = "os.tar.xz"
)

func untarGuestOS() {
	basePath := flagCielDir()
	parse()

	i := &ciel.Ciel{BasePath: *basePath}
	i.Check()
	c := i.Container()

	tar := flag.Arg(0)
	if tar == "" {
		d.SECTION("Download OS")
		d.ITEM("latest tarball url")
		d.Println(d.C(d.CYAN, LatestTarballURL))
		cmd := exec.Command("curl", "-o", DownloadTarball, "-#", LatestTarballURL)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		d.ITEM("download")
		if err != nil {
			d.FAILED_BECAUSE(err.Error())
			os.Remove(DownloadTarball)
			return
		}
		d.OK()
		tar = DownloadTarball
	}

	d.SECTION("Load OS From Compressed File")
	d.ITEM("are there any instances?")

	if instList := c.GetAllNames(); len(instList) != 0 {
		d.Println(d.C(d.YELLOW, strings.Join(instList, " ")))
		if d.ASK("DELETE ALL INSTANCES?", "yes/no") != "yes" {
			os.Exit(1)
		}
		for _, inst := range c.GetAll() {
			if inst.Running() {
				inst.Stop(context.TODO())
			}
			if inst.Mounted() {
				inst.Unmount()
			}
			d.ITEM("delete " + inst.Name)
			err := c.DelInst(inst.Name)
			d.ERR(err)
		}
	} else {
		d.Println(d.C(d.CYAN, "NO"))
	}

	d.ITEM("is dist dir empty?")
	os.Mkdir(c.DistDir(), 0755)
	list, err := ioutil.ReadDir(c.DistDir())
	if len(list) != 0 {
		d.Println(d.C(d.YELLOW, "NO"))
		if d.ASK("DELETE the old OS?", "yes/no") != "yes" {
			os.Exit(1)
		}
		d.ITEM("remove dist dir")
		if err := os.RemoveAll(c.DistDir()); err != nil {
			d.FAILED_BECAUSE(err.Error())
			os.Exit(1)
		}
		d.OK()

		d.ITEM("re-create dist dir")
		if err := os.Mkdir(c.DistDir(), 0755); err != nil {
			d.FAILED_BECAUSE(err.Error())
			os.Exit(1)
		}
		d.OK()
	} else if err != nil {
		d.FAILED_BECAUSE(err.Error())
		os.Exit(1)
	} else {
		d.Println(d.C(d.CYAN, "YES"))
	}

	d.ITEM("untar os")
	cmd := exec.Command("tar", "-xpf", tar, "-C", c.DistDir())
	output, err := cmd.CombinedOutput()
	if err != nil {
		d.FAILED_BECAUSE(strings.TrimSpace(string(output)))
	}
	d.OK()
}

func update() {
	var runErr error
	var exitStatus int
	defer func() {
		if runErr != nil {
			os.Exit(1)
		} else if exitStatus != 0 {
			os.Exit(exitStatus)
		}
	}()

	basePath := flagCielDir()
	networkFlag := flagNetwork()
	bootConfig := flagBootConfig()
	parse()

	i := &ciel.Ciel{BasePath: *basePath}
	i.Check()
	c := i.Container()

	d.SECTION("Update Guest Operating System")
	d.ITEM("are there online instances?")
	ready := true
	for _, inst := range c.GetAll() {
		if inst.Running() || inst.Mounted() {
			ready = false
			d.Print(d.C(d.YELLOW, inst.Name) + " ")
		}
	}
	if ready {
		d.Print(d.C(d.CYAN, "NO"))
	}
	d.Println()

	if !ready {
		if d.ASK("Stop all instances?", "yes/no") != "yes" {
			os.Exit(1)
		}
		for _, inst := range c.GetAll() {
			if inst.Running() {
				inst.Stop(context.TODO())
			}
			if inst.Mounted() {
				inst.Unmount()
			}
		}
	}

	const instName = "__update__"
	d.ITEM("create temporary instance")
	c.AddInst(instName)
	d.OK()
	defer func() {
		d.ITEM("delete temporary instance")
		c.DelInst(instName)
		d.OK()
	}()
	inst := c.Instance(instName)
	d.ITEM("mount temporary instance")
	inst.Mount()
	d.OK()
	defer func() {
		inst.Unmount()
	}()
	defer func() {
		inst.Stop(context.TODO())
	}()

	bootConf := strings.Split(strings.TrimSpace(*bootConfig), "\n")

	type ExitError struct{}
	var run = func(cmd string, poweroff bool) (int, error) {
		return _shellRun(inst, *networkFlag, true, bootConf, poweroff, cmd)
	}
	defer func() {
		p := recover()
		if p == nil {
			return
		}
		if _, isExit := p.(ExitError); !isExit {
			panic(p)
		}
		if runErr != nil {
			d.Println(d.C(d.RED, runErr.Error()))
		} else {
			d.Println(d.C(d.YELLOW, "INTERRUPTED, exit status: "+strconv.Itoa(exitStatus)))
		}
	}()

	exitStatus, runErr = run(`apt update --yes`, false)
	d.ITEM("update database")
	if runErr != nil || exitStatus != 0 {
		panic(ExitError{})
	}
	d.OK()

	exitStatus, runErr = run(`apt -o Dpkg::Options::="--force-confnew" full-upgrade --yes`, true)
	d.ITEM("update packages")
	if runErr != nil || exitStatus != 0 {
		panic(ExitError{})
	}
	d.OK()

	exitStatus, runErr = run(`apt autoremove --purge --yes`, true)
	d.ITEM("auto-remove packages")
	if runErr != nil || exitStatus != 0 {
		panic(ExitError{})
	}
	d.OK()

	d.ITEM("merge changes")
	err := inst.FileSystem().Merge()
	d.ERR(err)

	// TODO: clean
}
