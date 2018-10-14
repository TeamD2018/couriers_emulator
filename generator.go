package main

import (
	"context"
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
	return nil
}

func (g* Generator) CreateOrders() error {
	for _, w := range g.Workers {
		if err := w.CreateOrder(); err != nil {
			return err
		}
	}
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

func (g *Generator) UpdateWithInterval(interval time.Duration, throttle time.Duration, ctx context.Context) chan error {
	errchan := make(chan error)
	for _, w := range g.Workers {
		time.Sleep(throttle)
		go w.UpdateLocation(interval, ctx, errchan)
	}
	return errchan
}