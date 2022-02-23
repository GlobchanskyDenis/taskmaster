package main

import (
	"time"
	"fmt"
)

func main() {
	for {
		fmt.Println("Hello world")
		time.Sleep(time.Second * 2)
	}
}