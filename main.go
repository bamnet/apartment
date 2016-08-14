package main

import (
	"fmt"
	"log"

	"github.com/bamnet/apartment/wemo"
	"github.com/cenk/backoff"
)

func main() {
	water, err := wemo.NewDevice("192.168.1.187:49153")
	if err != nil {
		log.Fatalf("unable to connect to water: %v", err)
	}

	state, err := water.State()
	if err != nil {
		log.Fatalf("unable to get water state: %v", err)
	}
	fmt.Printf("Water state: %t\n", state)

	if err := backoff.Retry(func() error {
		return water.SetState(true)
	}, backoff.NewExponentialBackOff()); err != nil {
		log.Fatalf("unable to turn on water: %v", err)
	}
}
