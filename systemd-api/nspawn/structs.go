package nspawn

import "io"

type RunInfo struct {
	App    string
	Args   []string
	StdDev *StdDevInfo

	UseSystemdRun bool
}

type ContainerInfo struct {
	Init       bool
	InitArgs   []string
	Properties []string
	Network    *NetworkInfo
}

type StdDevInfo struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

// TODO: Network Configuration
// NOTE: NetworkInfo is not used for now.
type NetworkInfo struct {
	//Private     bool
	//IfaceMoveIn []string
	//MacVlan     []string
	//IpVlan      []string
	Zone string
}
