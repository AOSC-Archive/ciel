package main

import "flag"

func flagCielDir() *string {
	basePath := getEnv("CIEL_DIR", ".")
	flag.StringVar(&basePath, "d", basePath, "ciel work `directory`")
	return &basePath
}

func flagInstance() *string {
	instName := getEnv("CIEL_INST", "")
	flag.StringVar(&instName, "i", instName, "instance `name`")
	return &instName
}
