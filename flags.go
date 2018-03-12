package main

import (
	"flag"
	"os"
)

func flagCielDir() *string {
	basePath := getEnv("CIEL_DIR", ".")
	flag.StringVar(&basePath, "C", basePath, "ciel work `directory`; CIEL_DIR")
	return &basePath
}
func saveCielDir(basePath string) {
	saveEnv("CIEL_DIR", basePath)
}

func flagInstance() *string {
	instName := getEnv("CIEL_INST", "")
	flag.StringVar(&instName, "i", instName, "instance `name`; CIEL_INST")
	return &instName
}
func saveInstance(instName string) {
	saveEnv("CIEL_INST", instName)
}

func flagNetwork() *bool {
	network := getEnv("CIEL_NET", "false") == "true"
	flag.BoolVar(&network, "net", network, "create a network zone; CIEL_NET")
	return &network
}
func saveNetwork(network bool) {
	saveEnv("CIEL_NET", network)
}

func flagNoBooting() *bool {
	noBooting := getEnv("CIEL_BOOT", "false") == "true"
	flag.BoolVar(&noBooting, "n", noBooting, "do not boot the container; CIEL_BOOT")
	return &noBooting
}
func saveNoBooting(noBooting bool) {
	saveEnv("CIEL_BOOT", noBooting)
}

func flagArgs() *string {
	containerArg := getEnv("CIEL_FLAGS", "")
	return &containerArg
}
func saveArgs(args string) {
	saveEnv("CIEL_FLAGS", args)
}

func flagBatch() *bool {
	batch := getEnv("CIEL_BATCH_MODE", "false") == "true"
	flag.BoolVar(&batch, "batch", batch, "do not ask; CIEL_BATCH_MODE")
	return &batch
}
func saveBatch(batch bool) {
	saveEnv("CIEL_BATCH_MODE", batch)
}

func getEnv(key, def string) string {
	v, ok := os.LookupEnv(key)
	if !ok {
		return def
	}
	return v
}
func saveEnv(key string, value interface{}) {
	switch v := value.(type) {
	case bool:
		if v {
			os.Setenv(key, "true")
		} else {
			os.Setenv(key, "false")
		}
	case string:
		os.Setenv(key, v)
	}
}
