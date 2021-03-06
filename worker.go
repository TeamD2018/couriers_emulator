package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/icrowley/fake"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	moscowMinLon = 37.3468
	moscowMinLat = 55.5593
	moscowMaxLon = 37.8961
	moscowMaxLat = 55.9146
)

var ConstantAddresses = []Location{
	{
		Point: &GeoPoint{
			Lat: 55.797116,
			Lon: 37.537862,
		},
	},
	{
		Point: &GeoPoint{
			55.744584, 37.565937,
		},
	},
	{
		Point: &GeoPoint{
			55.765906, 37.683876,
		},
	},
	{
		Point: &GeoPoint{
			55.757219, 37.600293,
		},
	},
	{
		Point: &GeoPoint{
			55.707944, 37.683233,
		},
	},
}

var ConstantCouriers = []Courier{
	{Name: "Иван Васильев", Phone: "79039992231"},
	{Name: "Герман Стерлигов", Phone: "88005553535"},
	{Name: "Мстистлав Зеркальный", Phone: "78961235566"},
	{Name: "Константин Константинопольский", Phone: "79000011010"},
	{Name: "Борис Седых", Phone: "71127764388"},
	{Name: "Гавриил Степанов", Phone: "79110234578"},
	{Name: "Марат Ежов", Phone: "79167562345"},
	{Name: "Самуил Остапенко", Phone: "79991112222"},
	{Name: "Фарид Рахманов", Phone: "79458765134"},
	{Name: "Аскольд Щичко", Phone: "79901123476"},
}

var (
	AddrI = 0
	CourI = 0
)

type Geometry struct {
	Coordinates [][2]float64 `json:"coordinates"`
}

type Route struct {
	Geometry *Geometry `json:"geometry"`
	Legs     []*Leg    `json:"legs"`
}

type Annotation struct {
	Duration []float32 `json:"duration"`
}

type Leg struct {
	Annotation *Annotation `json:"annotation"`
}

type RouteResponse struct {
	Routes []*Route `json:"routes"`
}

type Worker struct {
	URL        string
	courier    *Courier
	orders     []*Order
	locations  []*GeoPoint
	durations  []float32
	iloc, idur int
	Interval   time.Duration
	client     *http.Client
	addrI      int
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
	err := fake.SetLang("ru")
	if err != nil {
		log.Printf("%s", err)
		return err
	}

	name := fake.FullName()
	phone := fake.Phone()
	var courier Courier
	if CourI < len(ConstantCouriers) {
		courier = ConstantCouriers[CourI]
		CourI++
	} else {
		courier = Courier{
			Name:  name,
			Phone: phone,
		}
	}
	res, _ := json.Marshal(courier)
	response, err := http.Post(w.buildURLCreate(w.URL), "application/json", bytes.NewReader(res))
	if err != nil {
		log.Printf("%s", err)
		return err
	}
	body, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
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

func (w *Worker) CreateOrder(routesURL string) error {
	locationSrc := w.getRandomLocation(moscowMinLat, moscowMaxLat, moscowMinLon, moscowMaxLon)
	locationDest := w.getRandomLocation(moscowMinLat, moscowMaxLat, moscowMinLon, moscowMaxLon)

	order := &Order{}
	order.OrderNumber = rand.Int() % 10000
	if AddrI < len(ConstantAddresses) {
		order.Destination = Location{
			Point: &GeoPoint{
				Lat: ConstantAddresses[AddrI].Point.Lat,
				Lon: ConstantAddresses[AddrI].Point.Lon,
			},
		}
		AddrI++
	} else {
		order.Destination = Location{
			Point: &GeoPoint{
				Lat: locationDest.Point.Lat,
				Lon: locationDest.Point.Lon,
			},
		}
	}

	if w.orders == nil || len(w.orders) == 0 {
		order.Source = Location{
			Point: &GeoPoint{
				Lat: locationSrc.Point.Lat,
				Lon: locationSrc.Point.Lon,
			},
		}
	} else {
		order.Source = w.orders[0].Source
	}

	res, _ := json.Marshal(order)

	response, err := http.Post(w.buildURLCreateOrder(w.URL), "application/json", bytes.NewReader(res))

	if err != nil {
		log.Printf("%s", err)
		return err
	}

	body, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		log.Printf("%s", err)
		return err
	}
	err = json.Unmarshal(body, order)
	if err != nil {
		log.Printf("%s", body)
		return err
	}

	w.orders = append(w.orders, order)

	return nil
}

func (w *Worker) DeleteCourier() error {
	req, err := http.NewRequest(http.MethodDelete, w.buildURLDelete(w.URL), nil)
	if err != nil {
		return err
	}
	resp, err := w.client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusNoContent {
		return errors.New(fmt.Sprintf("Not valid status code: expected 204, but %d", resp.StatusCode))
	}
	return nil
}

func (w *Worker) buildURLCreate(base string) string {
	return fmt.Sprintf("%s%s", base, "/couriers")
}

func (w *Worker) buildURLCreateOrder(base string) string {
	return fmt.Sprintf("%s/couriers/%s/orders", base, w.courier.ID)
}

func (w *Worker) buildURLUpdateOrder(base string, orderID string) string {
	return fmt.Sprintf("%s/couriers/%s/orders/%s", base, w.courier.ID, orderID)
}

func (w *Worker) buildURLUpdate(base string) string {
	return fmt.Sprintf("%s%s%s", base, "/couriers/", w.courier.ID)
}

func (w *Worker) buildURLGetRoute(routesURL string, locations []*Location) string {
	buf := strings.Builder{}
	buf.WriteString(fmt.Sprintf("%s/route/v1/driving/", routesURL))
	for i, l := range locations {
		buf.WriteString(fmt.Sprintf("%f,%f", l.Point.Lon, l.Point.Lat))
		if i != len(locations)-1 {
			buf.WriteByte(';')
		}
	}
	buf.WriteString("?geometries=geojson&annotations=duration&overview=full")
	return buf.String()
}

func (w *Worker) buildURLDelete(base string) string {
	return fmt.Sprintf("%s%s%s", base, "/couriers/", w.courier.ID)
}

func (w *Worker) getRoute(routesURL string) error {
	locations := make([]*Location, 0)
	if w.orders == nil {
		locations = append(locations, w.getRandomLocation(moscowMinLat, moscowMaxLat, moscowMinLon, moscowMaxLon))
		locations = append(locations, w.getRandomLocation(moscowMinLat, moscowMaxLat, moscowMinLon, moscowMaxLon))
	} else {
		locations = append(locations, &w.orders[0].Source)
		for _, o := range w.orders {
			locations = append(locations, &o.Destination)
		}
	}
	buildStr := w.buildURLGetRoute(routesURL, locations)

	log.Println(buildStr)

	response, err := http.Get(buildStr)

	if err != nil {
		log.Printf("%s", err)
		return err
	}

	b, _ := ioutil.ReadAll(response.Body)

	resp := RouteResponse{}

	err = json.Unmarshal(b, &resp)

	if err != nil {
		log.Printf("%s", b)
		panic(err)
	}

	log.Printf("%s\n", b)

	for _, r := range resp.Routes[0].Geometry.Coordinates {
		point := GeoPoint{Lon: r[0], Lat: r[1]}
		w.locations = append(w.locations, &point)
	}

	for _, l := range resp.Routes[0].Legs {
		for _, r := range l.Annotation.Duration {
			w.durations = append(w.durations, r)
		}
	}
	return nil
}

func (w *Worker) UpdateLocation(wg *sync.WaitGroup, speed int, routesURL string, interval time.Duration, ctx context.Context) {
	w.getRoute(routesURL)
	defer wg.Done()
	timer := time.NewTimer(time.Millisecond)
	for {
		if w.idur == len(w.durations) {
			log.Println("sending cancel orders...")
			w.SendCancelOrder()
			return
		}
		select {
		case <-timer.C:
			if err := w.update(); err != nil {
				return
			}
			timer.Reset(time.Duration((w.durations[w.idur]*1000)/float32(speed)) * time.Millisecond)
			w.idur++
		case <-ctx.Done():
			return
		}
	}
}

func (w *Worker) update() error {
	w.courier.Location = w.getNextLocationFromRoute()
	body, err := json.Marshal(w.courier)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPut, w.buildURLUpdate(w.URL), bytes.NewReader(body))
	resp, err := w.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (w *Worker) getNextLocationFromRoute() *Location {
	dlat, dlon := w.locations[w.iloc].Lat, w.locations[w.iloc].Lon
	if w.iloc < len(w.locations)-1 {
		w.iloc += 1
	}
	return &Location{&GeoPoint{dlat, dlon}}
}

func (w *Worker) getRandomLocation(minLat, maxLat, minLon, maxLon float64) *Location {
	return &Location{&GeoPoint{
		Lat: trim(minLat+rand.Float64()*(maxLat-minLat), 4),
		Lon: trim(minLon+rand.Float64()*(maxLon-minLon), 4),
	}}
}

func (w *Worker) SendCancelOrder() {
	for _, o := range w.orders {
		url := w.buildURLUpdateOrder(w.URL, o.ID)
		bodyStruct := struct {
			DeliveredAt int64 `json:"delivered_at"`
		}{
			time.Now().Unix(),
		}
		body, _ := json.Marshal(bodyStruct)
		if req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(body)); err != nil {
			log.Println(err)
		} else {
			if resp, err := w.client.Do(req); err != nil {
				log.Println(err)
			} else {
				if resp.StatusCode != http.StatusOK {
					log.Printf("error: expected %d, got %d", http.StatusOK, resp.StatusCode)
				}
			}
		}
	}
}

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func trim(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}
