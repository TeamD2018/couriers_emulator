package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	numCourier = flag.Int("couriers", 10, "Number of couriers to be created")
	timeout    = flag.Int("timeout", 60, "")
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
	if err := generator.CreateOrders(); err != nil {
		panic(err)
	}
	signalChan := make(chan os.Signal, 1)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(*timeout))
	doneGraceful := make(chan struct{})
	go func(ch chan struct{}) {
		signal.Notify(signalChan,
			syscall.SIGHUP,
			syscall.SIGINT,
			syscall.SIGTERM,
			syscall.SIGQUIT)
		select {
		case <-signalChan:
			cancel()
			fmt.Println("\nDeleting couriers...")
			if err := generator.DeleteCouriers(); err != nil {
				os.Exit(1)
			}
			os.Exit(0)
		case <-ch:
			return
		}
	}(doneGraceful)
	fmt.Printf("%d couriers created!\n", *numCourier)
	fmt.Println("Starting update locations...")

	ch := generator.UpdateWithInterval(
		time.Duration(*interval)*time.Second,
		time.Duration(*throttle)*time.Millisecond,
		ctx)
	if err := <-ch; err != nil {
		generator.DeleteCouriers()
		log.Println(err)
	} else {
		fmt.Println("\nDeleting couriers...")
		if err := generator.DeleteCouriers(); err != nil {
			doneGraceful <- struct{}{}
			os.Exit(1)
		}
		doneGraceful <- struct{}{}
	}
}
