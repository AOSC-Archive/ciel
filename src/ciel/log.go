package main

import (
	"io/ioutil"
	"log"
	"os"
)

const logFlags = log.Lshortfile | log.Ltime

var (
	errlog  = log.New(os.Stderr, "\033[31;1m[ERR ]\033[0m ", logFlags)
	warnlog = log.New(os.Stderr, "\033[33;1m[WARN]\033[0m ", logFlags)
	infolog = log.New(os.Stderr, "\033[32;1m[INFO]\033[0m ", logFlags)
	dbglog  = log.New(os.Stderr, "\033[39;1m[DBG ]\033[0m ", logFlags)
)

func SetLogLevel(logLevel int) {
	if logLevel < -2 {
		logLevel = -2
	} else if 2 <= logLevel {
		logLevel = 2
	}
	switch logLevel {
	case -2:
		errlog.SetOutput(ioutil.Discard)
		fallthrough
	case -1:
		warnlog.SetOutput(ioutil.Discard)
		fallthrough
	case 0:
		infolog.SetOutput(ioutil.Discard)
		fallthrough
	case 1:
		dbglog.SetOutput(ioutil.Discard)
		fallthrough
	case 2:
	}
}
