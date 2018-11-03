package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/pflag"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var (
	numCourier       *int
	timeout          *int64
	backend          *string
	url              *string
	interval         *int
	lag              *int
	ordersPerCourier *int
	routesURL        *string
	speed            *int
	mode             *string
)

func init() {
	numCourier = pflag.IntP("couriers", "c", 10, "number of couriers to be created")
	timeout = pflag.Int64P("timeout", "t", 60, "timeout for emulate courier activity")
	backend = pflag.StringP("backend", "b", "http://localhost:2015", "backend url")
	url = pflag.StringP("url", "u", "localhost:2018", "url of service emulator")
	interval = pflag.IntP("interval", "i", 5, "interval (in seconds) that couriers update geoposition")
	lag = pflag.IntP("lag", "l", 100, "lag (in milliseconds) that couriers update geoposition")
	ordersPerCourier = pflag.IntP("orders", "o", 1, "number of orders that can be affiliate to one courier")
	routesURL = pflag.StringP("routes", "r", "http://localhost:5000", "url for routes backend")
	speed = pflag.IntP("speed", "s", 1, "how fast the couriers traffic")
	mode = pflag.StringP("mode", "m", "s", "mode of emulator (c - console, s - server)")
	pflag.Parse()
	rand.Seed(time.Now().UTC().UnixNano())
}

func main() {
	if *mode == "s" {
		router := gin.Default()
		api := NewAPIService()
		router.POST("/test_data", api.GenerateTestData)
		router.DELETE("/test_data", api.DeleteCouriers)
		router.GET("/", api.GetHTML)
		if err := router.Run(*url); err != nil {
			return
		}
	}
	generator := NewGenerator(*backend, *numCourier)
	if err := generator.CreateCouriers(); err != nil {
		panic(err)
	}
	fmt.Printf("%d couriers created!\n", *numCourier)
	if err := generator.CreateOrders(*routesURL, rand.Intn(*ordersPerCourier)+1); err != nil {
		panic(err)
	}

	fmt.Printf("%d orders created (%d order for courier)\n", *numCourier*(*ordersPerCourier), *ordersPerCourier)

	doneChan, cancelChan, signalChan := make(chan struct{}), make(chan struct{}), make(chan os.Signal, 1)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(*timeout))

	go graceful(signalChan, cancel, cancelChan, doneChan, generator)

	fmt.Println("Starting update locations...")

	wg := &sync.WaitGroup{}

	generator.UpdateWithInterval(
		wg,
		*routesURL,
		*speed,
		time.Duration(*interval)*time.Second,
		time.Duration(*lag)*time.Millisecond,
		ctx)
	wg.Wait()
	signalChan <- syscall.SIGINT
	<-doneChan
}

func graceful(signalChan chan os.Signal, cancel context.CancelFunc, cancelChan chan struct{}, doneChan chan struct{}, generator *Generator) {
	signal.Notify(signalChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	for {
		select {
		case <-signalChan:
			cancel()
			fmt.Println("Deleting couriers...")
			if err := generator.DeleteCouriers(); err != nil {
				log.Println(err)
			}
			fmt.Println("Couriers deleted!")
			doneChan <- struct{}{}
			os.Exit(0)
		case <-cancelChan:
			return
		}
	}
}
