package proc

import (
	"bytes"
	"io/ioutil"
	"strconv"
	"strings"
)

func procPath(pid uint32) string {
	return "/proc/" + strconv.FormatUint(uint64(pid), 10)
}

func GetParentProcessID(pid uint32) (uint32, error) {
	b, err := ioutil.ReadFile(procPath(pid) + "/stat")
	if err != nil {
		return 0, err
	}
	dataSet := bytes.Split(b, []byte{'\x20'})
	ppid, err := strconv.ParseUint(string(dataSet[3]), 10, 32)
	return uint32(ppid), err
}

func GetCommandLineByPID(pid uint32) ([]string, error) {
	b, err := ioutil.ReadFile(procPath(pid) + "/cmdline")
	if err != nil {
		return nil, err
	}
	b = bytes.TrimSuffix(b, []byte{'\x00'})
	return strings.Split(string(b), string('\x00')), nil
}
