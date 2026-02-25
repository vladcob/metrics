package main

import (
	"net/http"
)

const (
	GaugeType   = "gauge"
	CounterType = "counter"
)

var Types = []string{
	GaugeType,
	CounterType,
}

type MemStorageInterface interface {
	addCounter(name string, value int64)
	addGauge(name string, value float64)
	updateCounter(name string, value int64)
	updateGauge(name string, value float64)
	removeCounter(name string)
	removeGauge(name string)
}

type MemStorage struct {
	counters map[string]int64
	gauges   map[string]float64
}

type Middleware func(http.Handler) http.Handler

func main() {
	http.Handle("/update/{type}/{name}/{value}", Conveyor(http.HandlerFunc(updateMetric), notFoundMiddleware, badRequestMiddleware))

	err := http.ListenAndServe(`:8080`, nil)

	if err != nil {
		panic(err)
	}
}

func updateMetric(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Ok status"))
}

func Conveyor(h http.Handler, middlewares ...Middleware) http.Handler {
	for _, middleware := range middlewares {
		h = middleware(h)
	}
	return h
}

func notFoundMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if name := r.PathValue("name"); name == "" {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Metric name not found"))
		}

		next.ServeHTTP(w, r)
	})
}

func badRequestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isValid := false

		for _, v := range Types {
			if v == r.PathValue("type") {
				isValid = true
			}
		}

		if isValid {
			next.ServeHTTP(w, r)
		} else {

			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Invalid metric type"))
		}
	})
}
