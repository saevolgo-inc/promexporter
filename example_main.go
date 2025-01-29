package main

import (
	"fmt"
	"log/slog"
	"math/rand/v2"
	POM "https://github.com/saevolgo-inc/promexporter"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/kaihendry/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	Version   string
	GoVersion = runtime.Version()
)

func main() {
	slog.SetDefault(getLogger(os.Getenv("LOGLEVEL")))

	mymetric := POM.NewCounterVecMultiLabels("companyNameSpace", "devs", "We are testing this", []POM.Labels{POM.Labels{Name: "method", Value: "1"}, POM.Labels{Name: "type", Value: "2"}})
	mymetric2 := POM.NewCounterVecMultiLabels("companyNameSpace", "prods", "Production Metric", []POM.Labels{POM.Labels{Name: "method", Value: "1"}, POM.Labels{Name: "type", Value: "2"}, POM.Labels{Name: "elevation", Value: "2"}})

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	go RandomMetricGenerator3(mymetric2)
	go RandomMetricGenerator(mymetric)

	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = "8085"
	}
	if _, err := strconv.Atoi(port); err != nil {
		slog.Error("invalid port", "port", port, "error", err)
		os.Exit(1)
	}

	slog.Info("starting slo", "port", port, "Version", Version, "GoVersion", GoVersion)

	if err := http.ListenAndServe(":"+port, middleware.LogStatus(mux)); err != nil {
		slog.Error("error listening", "error", err)
	}
}
func RandomMetricGenerator(mymetric *POM.MetricMetadata) {
	for {
		time.Sleep(3 * time.Second)
		m1 := rand.IntN(5)
		m2 := rand.IntN(5)
		x1 := strconv.Itoa(m1)
		x2 := strconv.Itoa(m2)
		mymetric.IncrementCounterVecMultiLabelValuesOnly(x1, x2)
	}
}
func RandomMetricGenerator3(mymetric *POM.MetricMetadata) {
	for {
		time.Sleep(3 * time.Second)
		m1 := rand.IntN(5)
		m2 := rand.IntN(5)
		m3 := rand.IntN(3)
		x1 := strconv.Itoa(m1)
		x2 := strconv.Itoa(m2)
		x3 := strconv.Itoa(m3)
		mymetric.IncrementCounterVecMultiLabel([]POM.Labels{POM.Labels{Name: "method", Value: x1}, POM.Labels{Name: "type", Value: x2}, POM.Labels{Name: "elevation", Value: x3}})
	}
}
func getLogger(logLevel string) *slog.Logger {
	levelVar := slog.LevelVar{}

	if logLevel != "" {
		if err := levelVar.UnmarshalText([]byte(logLevel)); err != nil {
			panic(fmt.Sprintf("Invalid log level %s: %v", logLevel, err))
		}
	}

	return slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: levelVar.Level(),
	}))
}
