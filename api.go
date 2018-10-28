package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"
)

type APIService struct {
	generator            *Generator
	cancel               context.CancelFunc
	cancelChan, doneChan chan struct{}
	wg                   *sync.WaitGroup
}

func NewAPIService() *APIService {
	return &APIService{cancelChan: make(chan struct{}), doneChan: make(chan struct{}, 1)}
}

func (api *APIService) DeleteCouriers(ctx *gin.Context) {
	api.cancel()
	api.wg.Wait()
	if err := api.generator.DeleteCouriers(); err != nil {
		log.Println(err)
	}
	api.cancelChan <- struct{}{}
	ctx.Status(http.StatusOK)
}

func (api *APIService) GenerateTestData(ctx *gin.Context) {
	if api.cancel != nil {
		api.cancel()
		api.wg.Wait()
		if err := api.generator.DeleteCouriers(); err != nil {
			log.Println(err)
		}
		api.cancelChan <- struct{}{}
	}

	api.generator = NewGenerator(*backend, *numCourier)

	if err := api.generator.CreateCouriers(); err != nil {
		log.Println(err)
	}

	c, cancel := context.WithCancel(context.Background())

	api.cancel = cancel

	go graceful(make(chan os.Signal, 1), api.cancel, api.cancelChan, api.doneChan, api.generator)

	log.Printf("%d couriers created!\n", *numCourier)

	if err := api.generator.CreateOrders(*routesURL, int(rand.Int31n(int32(*ordersPerCourier)))); err != nil {
		log.Println(err)
	}

	log.Printf("%d orders created (%d order for courier)\n", *numCourier*(*ordersPerCourier), *ordersPerCourier)


	log.Println("Starting update locations...")

	wg := &sync.WaitGroup{}

	api.wg = wg

	api.generator.UpdateWithInterval(
		wg,
		*routesURL,
		*speed,
		time.Duration(*interval)*time.Second,
		time.Duration(*lag)*time.Millisecond,
		c)
}
