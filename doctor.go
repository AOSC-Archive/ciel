package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"ciel/internal/ciel"
	"ciel/ipc"
	"ciel/proc-api"
	"ciel/systemd-api/machined"
	"fmt"
)

func dbus_test() {
	m := machined.NewManager()
	machine, err := m.GetMachine("meow_12000e")
	if err != nil {
		log.Fatalln(err)
	}
	leader, err := machine.Leader()
	if err != nil {
		log.Fatalln(err)
	}
	host, err := proc.GetParentProcessID(leader)
	if err != nil {
		log.Fatalln(err)
	}
	cmdline, _ := proc.GetCommandLineByPID(host)
	for _, arg := range cmdline {
		if arg == "-b" || arg == "--boot" {
			//return True
			fmt.Println("boot")
			return
		}
	}
	fmt.Println("non-boot")
}

func doctor() {
	dbus_test()
	return
	ppid, _ := proc.GetParentProcessID(3238)
	cmdline, _ := proc.GetCommandLineByPID(ppid)
	for _, arg := range cmdline {
		if arg == "-b" || arg == "--boot" {
			//return True
			break
		}
	}
	//return False
	fmt.Println(cmdline)
	return
	// TODO: doctor
	//checkUname()
	//checkOverlayfs()
	//checkSystemd()
	//checkTar()
	basePath := flagCielDir()
	instName := flagInstance()
	networkFlag := flagNetwork()
	noBooting := flagNoBooting()
	args := flagArgs()
	parse()

	if flag.NArg() > 1 {
		log.Fatalln("you must pass one argument only")
	}

	//shmKey := sync.GenerateKey(*basePath, 21)
	//shmId := sync.GetSharedMemory(shmKey, 12)
	//println("shmid", sync.AttachSharedMemory(shmId))
	l := ipc.NewMutex(*basePath, 12, true)
	if !l.TryLock() {
		log.Fatalln("cannot run more than one instance.")
	}
	fmt.Println("ohayo")
	fmt.Scanln()
	return

	i := &ciel.Ciel{BasePath: *basePath}
	i.Check()
	c := i.Container()
	c.CheckInst(*instName)

	inst := c.Instance(*instName)
	inst.Mount()

	containerArgs := strings.Split(strings.TrimSpace(*args), "\n")

	if flag.NArg() == 0 {
		exitStatus, err := _openShell(
			inst,
			*networkFlag,
			!*noBooting,
			containerArgs,
		)
		if err != nil {
			log.Println(err)
		}
		os.Exit(exitStatus)
	}
	exitStatus, err := _shellRun(
		inst,
		*networkFlag,
		!*noBooting,
		containerArgs,
		false,
		flag.Arg(0),
	)
	if err != nil {
		log.Println(err)
	}
	os.Exit(exitStatus)
}

//func checkUname() {
//	uname := syscall.Utsname{}
//	err := syscall.Uname(&uname)
//}
