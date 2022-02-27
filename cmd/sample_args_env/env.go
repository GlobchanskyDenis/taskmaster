package main

import (
	"os"
)

var gEnv *string

func parseEnv(envName string) {
    env, exists := os.LookupEnv(envName)
	if exists == true {
		gEnv = &env
	}
}