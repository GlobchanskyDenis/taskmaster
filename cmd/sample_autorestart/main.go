package main

import (
	"time"
	"os"
)

func main() {
	println("Стартую программу")
	time.Sleep(time.Millisecond * 1000)
	println("Подождал секунду, можно завершать работу программы")
	os.Exit(100)
}