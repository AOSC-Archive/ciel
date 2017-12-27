package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"ciel/internal/ciel"
	"ciel/internal/display"
	"io/ioutil"
)

func untarGuestOS() {
	basePath := flagCielDir()
	parse()

	i := &ciel.Ciel{BasePath: *basePath}
	i.Check()
	c := i.Container()

	if tar := flag.Arg(0); tar == "" {
		log.Fatalln("no tar file specified")
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
				d.ITEM("shutdown " + inst.Name)
				err := inst.Stop(context.TODO())
				d.ERR(err)
			}
			if inst.Mounted() {
				d.ITEM("unmount " + inst.Name)
				err := inst.Unmount()
				d.ERR(err)
			}
			d.ITEM("delete " + inst.Name)
			err := c.InstDel(inst.Name)
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
	cmd := exec.Command("tar", "-xpf", flag.Arg(0), "-C", c.DistDir())
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
			d.Print(d.C(d.RED, inst.Name) + " ")
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
				d.ITEM("shutdown " + inst.Name)
				err := inst.Stop(context.TODO())
				d.ERR(err)
			}
			if inst.Mounted() {
				d.ITEM("unmount " + inst.Name)
				err := inst.Unmount()
				d.ERR(err)
			}
		}
	}

	const instName = "__update__"
	d.ITEM("create temporary instance")
	c.InstAdd(instName)
	d.OK()
	defer func() {
		d.ITEM("delete temporary instance")
		c.InstDel(instName)
		d.OK()
	}()
	inst := c.Instance(instName)
	d.ITEM("mount temporary instance")
	inst.Mount()
	d.OK()
	defer func() {
		d.ITEM("unmount temporary instance")
		inst.Unmount()
		d.OK()
	}()
	defer func() {
		d.ITEM("stop temporary instance")
		inst.Stop(context.TODO())
		d.OK()
	}()

	bootConf := strings.Split(strings.TrimSpace(*bootConfig), "\n")
	shell, err := inst.Shell("root")
	if err != nil {
		log.Fatal(err)
	}
	args := []string{
		shell,
		"--login",
		"-c",
	}
	ctx := context.TODO()

	type ExitError struct{}
	var run = func(cmd string) (int, error) {
		return inst.Run(ctx, true, *networkFlag, bootConf, append(args, cmd)...)
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

	exitStatus, runErr = run(`apt update --yes`)
	d.ITEM("update database")
	if runErr != nil || exitStatus != 0 {
		panic(ExitError{})
	}
	d.OK()

	exitStatus, runErr = run(`apt -o Dpkg::Options::="--force-confnew" full-upgrade --yes`)
	d.ITEM("update packages")
	if runErr != nil || exitStatus != 0 {
		panic(ExitError{})
	}
	d.OK()

	exitStatus, runErr = run(`apt autoremove --purge --yes`)
	d.ITEM("auto-remove packages")
	if runErr != nil || exitStatus != 0 {
		panic(ExitError{})
	}
	d.OK()

	d.ITEM("merge changes")
	err = inst.FileSystem().Merge()
	d.ERR(err)

	// TODO: clean
}
