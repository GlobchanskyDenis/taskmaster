package main

import (
	"flag"
	"fmt"
)

var gFlag *string

func parseFlags() {
	var arg string
	flag.StringVar(&arg, "X", "", "Аргумент который будет выведен")

	flag.Parse()

	if arg != "" {
		gFlag = &arg
	}
}

func parseFlags_deprecated() {
	// var arg string
	flagVar := flag.Lookup("x")

	flag.Parse()

	if flagVar != nil {
		fmt.Printf("flag x %#v\n", flagVar)
		val := flagVar.Value.String()
		gFlag = &val
	}

	// if arg != "" {
	// 	gFlag = &arg
	// }
}