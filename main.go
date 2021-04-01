package promexporter

import (
	"errors"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	Counters = make(map[string]counterData)
	Gauges   = make(map[string]gaugeData)
)

type counterData struct {
	Counter    prometheus.Counter
	IncChannel chan int
}

type gaugeData struct {
	Gauge      prometheus.Gauge
	ValChannel chan float64
}

type MetricMetadata struct {
	Namespace string
	Name      string
	Help      string
}

// SetupGauge - sets the value of gauge to the one provided
func SetupGauge(namespace, id, help string, val float64) (bool, error) {
	gauge, ok := Gauges[id]
	if ok {
		gauge.ValChannel <- val
		return true, nil
	} else {
		CreateGauge(id, MetricMetadata{
			Name:      id,
			Namespace: namespace,
			Help:      help,
		})
	}
	gauge, ok = Gauges[id]
	if ok {
		gauge.ValChannel <- val
		return true, nil
	} else {
		return false, errors.New("[SetupGauge] existing gauge not found | failed to create new gauge")
	}
}

// Increments Value of Counter by 1 on given counter id
func IncrementCounter(namespace, id, help string) (bool, error) {
	counter, ok := Counters[id]
	if ok {
		counter.IncChannel <- 1
		return true, nil
	} else {
		CreateCounter(id, MetricMetadata{
			Name:      id,
			Namespace: namespace,
			Help:      help,
		})
	}
	counter, ok = Counters[id]
	if ok {
		counter.IncChannel <- 1
		return true, nil
	} else {
		return false, errors.New("[IncrementCounter] existing counter not found | failed to create new counter")
	}
}

// Creates Counter based on the supplied metrics metadata
func CreateCounter(id string, data MetricMetadata) {
	Counters[id] = counterData{
		Counter: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: data.Namespace,
			Name:      data.Name,
			Help:      data.Help,
		}),
		IncChannel: make(chan int),
	}
	prometheus.MustRegister(Counters[id].Counter)
	go func(cd counterData) {
		for {
			<-cd.IncChannel
			cd.Counter.Inc()
		}
	}(Counters[id])
}

// Creates Counters based on the supplied metrics metadata
func CreateCounters(data map[string]MetricMetadata) {
	for k, v := range data {
		Counters[k] = counterData{
			Counter: prometheus.NewCounter(prometheus.CounterOpts{
				Namespace: v.Namespace,
				Name:      v.Name,
				Help:      v.Help,
			}),
			IncChannel: make(chan int),
		}
	}
}

// Creates Gaguge based on the supplied metrics metadata
func CreateGauge(id string, data MetricMetadata) {
	Gauges[id] = gaugeData{
		Gauge: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: data.Namespace,
			Name:      data.Name,
			Help:      data.Help,
		}),
		ValChannel: make(chan float64),
	}
	prometheus.MustRegister(Gauges[id].Gauge)
	go func(cd gaugeData) {
		for {
			val := <-cd.ValChannel
			cd.Gauge.Set(val)
		}
	}(Gauges[id])
}

// Creates Gauges based on the supplied metrics metadata
func CreateGauges(data map[string]MetricMetadata) {
	for k, v := range data {
		Gauges[k] = gaugeData{
			Gauge: prometheus.NewGauge(prometheus.GaugeOpts{
				Namespace: v.Namespace,
				Name:      v.Name,
				Help:      v.Help,
			}),
			ValChannel: make(chan float64),
		}
	}
}

// Registers Metrics route
func RegisterRoute(r *mux.Router) {
	r.Path("/metrics").Handler(promhttp.Handler())
}

// Registers newly created metrics
func Register() {
	registerCounters()
	registerGauges()
}

func StartOps() {
	for _, v := range Counters {
		go func(cd counterData) {
			for {
				<-cd.IncChannel
				cd.Counter.Inc()
			}
		}(v)
	}
	for _, v := range Gauges {
		go func(cd gaugeData) {
			for {
				val := <-cd.ValChannel
				cd.Gauge.Set(val)
			}
		}(v)
	}
}

func registerCounters() {
	for _, v := range Counters {
		prometheus.MustRegister(v.Counter)
	}
}

func registerGauges() {
	for _, v := range Gauges {
		prometheus.MustRegister(v.Gauge)
	}
}
