package prometheus

import (
	"errors"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	Counters  = make(map[string]counterData)
	Gauges    = make(map[string]gaugeData)
	GaugeVecs = make(map[string]guagevecData)
)

type counterData struct {
	Counter    prometheus.Counter
	IncChannel chan int
}

type gaugeData struct {
	Gauge      prometheus.Gauge
	ValChannel chan float64
}
type guagevecData struct {
	Guagevec   *prometheus.GaugeVec
	ValChannel chan GuageVecMetric
}
type Labels struct {
	Name  string
	Value string
}
type GuageVecMetric struct {
	MetricValue float64
	MetricName  string
	LabelInfo   []Labels
}
type MetricMetadata struct {
	Namespace  string
	Name       string
	LabelTitle string
	LabelVal   string
	Help       string
	LabelInfo  []Labels
}

// SetupGaugeVecWithMultiLabels - sets the value of gauge to the one provided
func SetupGaugeVecWithMultiLabels(namespace, id, help string, lbls []Labels, val float64) (bool, error) {

	gauge, ok := GaugeVecs[id]
	if ok {
		gauge.Guagevec.Reset()
		gauge.ValChannel <- GuageVecMetric{MetricValue: val, LabelInfo: lbls}
		return true, nil
	} else {
		CreateGaugeVecWithMultiLabels(id, MetricMetadata{
			Name:      id,
			Help:      help,
			LabelInfo: lbls,
		})
	}
	gauge, ok = GaugeVecs[id]
	if ok {
		gauge.ValChannel <- GuageVecMetric{MetricValue: val, LabelInfo: lbls}
		return true, nil
	} else {
		return false, errors.New("[SetupGauge] existing gauge not found | failed to create new gauge")
	}
}

// SetupGaugeVec - sets the value of gauge to the one provided
func SetupGaugeVec(namespace, id, help, labelTitle, labelVal string, val float64) (bool, error) {

	gauge, ok := GaugeVecs[id]
	if ok {
		gauge.ValChannel <- GuageVecMetric{MetricValue: val, MetricName: labelVal}
		return true, nil
	} else {
		CreateGaugeVec(id, MetricMetadata{
			Name:       id,
			Help:       help,
			LabelTitle: labelTitle,
			LabelVal:   labelVal,
		})
	}
	gauge, ok = GaugeVecs[id]
	if ok {
		gauge.ValChannel <- GuageVecMetric{MetricValue: val, MetricName: labelVal}
		return true, nil
	} else {
		return false, errors.New("[SetupGauge] existing gauge not found | failed to create new gauge")
	}
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

// CreateGaugeVecWithMultiLabels - Creates Gaguge based on the supplied metrics metadata
func CreateGaugeVecWithMultiLabels(id string, data MetricMetadata) {
	var definedLabels []string
	for _, labels := range data.LabelInfo {
		definedLabels = append(definedLabels, labels.Name)
	}
	GaugeVecs[id] = guagevecData{
		Guagevec: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: data.Namespace,
			Name:      data.Name,
			Help:      data.Help,
		},
			definedLabels),
		ValChannel: make(chan GuageVecMetric),
	}

	prometheus.MustRegister(GaugeVecs[id].Guagevec)
	go func(cd guagevecData) {
		for {
			val := <-cd.ValChannel
			var labelsGiven = make(map[string]string)
			for _, x := range val.LabelInfo {
				labelsGiven[x.Name] = x.Value
			}
			cd.Guagevec.With(labelsGiven).Set(float64(val.MetricValue))
		}
	}(GaugeVecs[id])
}

// Creates Gaguge based on the supplied metrics metadata
func CreateGaugeVec(id string, data MetricMetadata) {
	GaugeVecs[id] = guagevecData{
		Guagevec: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: data.Namespace,
			Name:      data.Name,
			Help:      data.Help,
		},
			[]string{data.LabelTitle}),
		ValChannel: make(chan GuageVecMetric),
	}

	prometheus.MustRegister(GaugeVecs[id].Guagevec)
	go func(cd guagevecData) {
		for {
			val := <-cd.ValChannel
			cd.Guagevec.WithLabelValues(val.MetricName).Set(float64(val.MetricValue))
		}
	}(GaugeVecs[id])
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
	registerGaugeVecs()
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
func registerGaugeVecs() {
	for _, v := range GaugeVecs {
		prometheus.MustRegister(v.Guagevec)
	}
}
