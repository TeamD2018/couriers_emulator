package main

import (
	"flag"
	"fmt"
	"log"
	"time"
)

var (
	numCourier = flag.Int("couriers", 10, "Number of couriers to be created")
	url        = flag.String("url", "http://localhost:2015", "")
	interval   = flag.Int("interval", 5, "interval of update geoposition by courier")
	throttle   = flag.Int("throttle", 500, "throttle")
)

func main() {
	flag.Parse()
	generator := NewGenerator(*url, *numCourier)
	if err := generator.CreateCouriers(); err != nil {
		panic(err)
	}
	fmt.Printf("%d couriers created!\n", *numCourier)
	fmt.Println("Starting update locations...")
	ch := generator.UpdateWithInterval(time.Duration(*interval)*time.Second, time.Duration(*throttle)*time.Millisecond)
	err := <-ch
	log.Println(err)
}
