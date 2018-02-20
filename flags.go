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

func flagBootConfig() *string {
	bootConfig := getEnv("CIEL_BOOTCFG", "")
	return &bootConfig
}
func saveBootConfig(bootConfig string) {
	saveEnv("CIEL_BOOTCFG", bootConfig)
}

func flagCI() *bool {
	ci := getEnv("CIEL_CI_MODE", "false") == "true"
	flag.BoolVar(&ci, "ci", ci, "do not ask; CIEL_CI_MODE")
	return &ci
}
func saveCI(network bool) {
	saveEnv("CIEL_CI_MODE", network)
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
