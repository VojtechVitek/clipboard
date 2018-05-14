package main

import (
	"fmt"
	"log"
	"time"

	"github.com/VojtechVitek/clipboard"
)

func loop() {
	lastValue := ""
	for {
		value, err := clipboard.Get()
		if err != nil {
			log.Println(err)
			time.Sleep(1 * time.Second)
			continue
		}

		if value != lastValue {
			fmt.Println(value)
		}
		lastValue = value

		time.Sleep(1 * time.Second)
	}
}

func main() {
	go loop()

	// TODO: UI.
	select {}
}
