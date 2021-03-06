package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"html/template"
	"log"
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

func (api *APIService) GetHTML(ctx *gin.Context) {
	t, err := template.New("web.html").ParseFiles("web.html")
	if err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	var r struct {
		OrderID   string
		CourierID string
	}

	if api.generator == nil {
		r.OrderID = "0"
		r.CourierID = "0"
	} else {
		for i := 0; i < len(api.generator.Workers); i++ {
			if api.generator.Workers[i].orders != nil {
				r.OrderID = api.generator.Workers[i].orders[0].ID
				r.CourierID = api.generator.Workers[i].courier.ID
				break
			}
		}
	}

	if err := t.Execute(ctx.Writer, r); err != nil {
		log.Println(err)
		ctx.Status(http.StatusInternalServerError)
	}
	ctx.Status(http.StatusOK)
}

func (api *APIService) DeleteCouriers(ctx *gin.Context) {
	api.cancel()
	api.cancel = nil
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
		api.cancel = nil
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

	if err := api.generator.CreateOrders(*routesURL, *ordersPerCourier); err != nil {
		log.Println(err)
	}

	ctx.AbortWithStatus(http.StatusOK)

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
