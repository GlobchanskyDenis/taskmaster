package main

import (
	"fmt"
)

const ENV_NAME = "TEMP_ENV"

func printAll() {
	if gFlag == nil {
		fmt.Printf("Флаг X не был задан\n")
	} else {
		fmt.Printf("Флаг X == %s\n", *gFlag)
	}

	if gEnv == nil {
		fmt.Printf("Переменная окружения %s не была задана\n", ENV_NAME)
	} else {
		fmt.Printf("Переменная окружения %s == %s\n", ENV_NAME, *gEnv)
	}
}

func main() {
	parseFlags()
	parseEnv(ENV_NAME)
	printAll()

	// setUmask(0o77)
	if err := createTestFile("test1.txt", "1"); err != nil {
		println(err.Error())
	}
	// setUmask(0o00)
	if err := createTestFile("test2.txt", "2"); err != nil {
		println(err.Error())
	}
}