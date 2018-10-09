package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/icrowley/fake"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net/http"
	"time"
)

type Worker struct {
	URL      string
	courier  *Courier
	Interval time.Duration
	client   *http.Client
}

func NewWorker(url string) *Worker {
	client := &http.Client{}
	return &Worker{
		courier: &Courier{},
		client:  client,
		URL:     url,
	}
}

func (w *Worker) CreateCourier() error {
	name := fake.MaleFullName()
	phone := fake.Phone()
	courier := Courier{
		Name:  name,
		Phone: &phone,
	}
	res, _ := json.Marshal(courier)
	response, err := http.Post(w.buildURLCreate(w.URL), "application/json", bytes.NewReader(res))
	if err != nil {
		log.Printf("%s", err)
		return err
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Printf("%s", err)
		return err
	}
	err = json.Unmarshal(body, w.courier)
	if err != nil {
		log.Printf("%s", body)
		return err
	}
	return nil
}

func (w *Worker) buildURLCreate(base string) string {
	return fmt.Sprintf("%s%s", base, "/couriers")
}

func (w *Worker) buildURLUpdate(base string) string {
	return fmt.Sprintf("%s%s%s", base, "/couriers/", w.courier.ID)
}

func (w *Worker) UpdateLocation(interval time.Duration, errchan chan<- error) {
	w.Interval = interval
	for {
		//garbage collector sucks
		<-time.After(w.Interval)
		if err := w.update(); err != nil {
			errchan <- err
		}
	}
}

func (w *Worker) update() error {
	w.courier.Location = diffLocation(w.courier.Location)
	body, err := json.Marshal(w.courier)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPut, w.buildURLUpdate(w.URL), bytes.NewReader(body))
	_, err = w.client.Do(req)
	if err != nil {
		return err
	}
	return nil
}

func diffLocation(prevLocation *Location) *Location {
	if prevLocation == nil {
		return &Location{&GeoPoint{
			Lat: trim(-90.0+rand.Float64()*180.0, 4),
			Lon: trim(-180.0+rand.Float64()*360.0, 4),
		}}
	}
	dlat, dlon := trim(-0.0005+rand.Float64()*0.001, 4), trim(-0.0005+rand.Float64()*0.001, 4)
	if dlat < -90.0 {
		dlat = -90.0
	}
	if dlat > 90.0 {
		dlat = 90
	}
	if dlon < -180.0 {
		dlon = -180.0
	}
	if dlon > 180.0 {
		dlon = 180.0
	}
	prevLocation.Point.Lat += dlat
	prevLocation.Point.Lon += dlon
	return prevLocation
}

func trim(n float64, precision int) float64 {
	intn := int(n * math.Pow10(precision))
	return float64(intn) / math.Pow10(precision)
}
