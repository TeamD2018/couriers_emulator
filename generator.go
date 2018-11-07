package main

import (
	"context"
	"math/rand"
	"sync"
	"time"
)

type Generator struct {
	Workers []*Worker
}

func NewGenerator(url string, numWorkers int) *Generator {
	g := &Generator{}
	for i := 0; i < numWorkers; i++ {
		g.Workers = append(g.Workers, NewWorker(url))
	}
	return g
}

func (g *Generator) CreateCouriers() error {
	for _, w := range g.Workers {
		if err := w.CreateCourier(); err != nil {
			return err
		}
	}
	CourI = 0
	return nil
}

func (g *Generator) CreateOrders(routesURL string, numberOfOrders int) error {
	for _, w := range g.Workers {
		for i := 0; i < rand.Intn(numberOfOrders); i++ {
			if err := w.CreateOrder(routesURL); err != nil {
				return err
			}
		}
	}
	AddrI = 0
	return nil
}

func (g *Generator) DeleteCouriers() error {
	for _, w := range g.Workers {
		if err := w.DeleteCourier(); err != nil {
			return err
		}
	}
	return nil
}

func (g *Generator) UpdateWithInterval(wg *sync.WaitGroup, routeURL string, speed int, interval, throttle time.Duration, ctx context.Context) {
	for _, w := range g.Workers {
		wg.Add(1)
		go w.UpdateLocation(wg, speed, routeURL, interval, ctx)
		time.Sleep(throttle)
	}
}
