package main

import "time"

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

func (g *Generator) UpdateWithInterval(interval time.Duration, throttle time.Duration) chan error {
	errchan := make(chan error)
	for _, w := range g.Workers {
		time.Sleep(throttle)
		go w.UpdateLocation(interval, errchan)
	}
	return errchan
}