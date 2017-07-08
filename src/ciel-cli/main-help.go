package main

import (
	"fmt"
	"os"
)

func printHelp() {
	fmt.Println("Usage: " + os.Args[0] + " [command [arg ...]]")
	fmt.Println(`Default command is "shell".`)
	fmt.Println("")

	fmt.Println("Commands:")
	fmt.Println("\tinit <tar>\tinit and extract \"stub\" tarball")
	fmt.Println("\tstub-upd  \tupdate all packages for stub, reset dist")
	// fmt.Println("\tstub-release\tbuild tarball for stub")
	fmt.Println("\tdist-cfg  \tinstall packages, etc.")
	// fmt.Println("\tdist-upd  \tupdate all packages for dist")
	fmt.Println("\tshell     \tlaunch shell (bash)")
	// fmt.Println("\tmount       \tmount, run a shell ON HOST, unmount")

	fmt.Println("\thelp      \tshow this message")
}
