package main

import (
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

//自定义结构体
type CpuCollector struct {
	cpuDesc *prometheus.Desc
}

func NewCpuCollector() *CpuCollector {
	return &CpuCollector{
		cpuDesc: prometheus.NewDesc(
			"test_cpu_percent_v3",
			"CPU Perent V3",
			[]string{"cpu"},
			nil,
		),
	}
}

func (c *CpuCollector) Describe(descs chan<- *prometheus.Desc) {
	descs <- c.cpuDesc
}

func (c *CpuCollector) Collect(metrics chan<- prometheus.Metric) {
	for i := 0; i < 4; i++ {
		metrics <- prometheus.MustNewConstMetric(
			c.cpuDesc,
			prometheus.GaugeValue,
			rand.Float64(),
			strconv.Itoa(i),
		)
	}
}

func main() {
	addr := ":9090"
	// 定义指标：类型 有标签/无标签
	// 固定lable
	totalV1 := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace:   "",
			Subsystem:   "",
			Name:        "test_total_v1",
			Help:        "Test Total V1 Counter",
			ConstLabels: map[string]string{"name": "v1"},
		},
	)

	totalV2 := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace:   "test",
			Subsystem:   "total",
			Name:        "v2",
			Help:        "Test Total V2 Counter",
			ConstLabels: prometheus.Labels{"name": "v2"},
		},
		[]string{"path"},
	)

	totalV3 := prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Name:        "test_total_v3",
			Help:        "Test Total V3 Counter",
			ConstLabels: prometheus.Labels{"name": "v3"},
		},
		func() float64 {
			return rand.Float64()
		},
	)

	cpuPercent := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "test_cpu_percent",
			Help: "Test CPU PerCent Guage",
		},
		[]string{"cpu"},
	)

	requestTime := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "request_time",
			Help: "Request Time Histogram",
			Buckets: prometheus.LinearBuckets(0, 3, 3),
		},
		[]string{"path"},	
	)

	requestTimeS := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "request_time_s",
			Help: "Request Time Summary",
			Objectives: map[float64]float64{0.9:0.01, 0.8:0.02, 0.7:0.03, 0.6:0.05},
		},
		[]string{"path"},
	)

	// 注册指标
	prometheus.MustRegister(totalV1)
	prometheus.MustRegister(totalV2)
	prometheus.MustRegister(totalV3)
	prometheus.MustRegister(cpuPercent)
	prometheus.MustRegister(requestTime)
	prometheus.MustRegister(requestTimeS)
	prometheus.MustRegister(NewCpuCollector())
	// 更新指标采样值 => 方法
	go func() {
		<-time.After(10 * time.Second)
		totalV1.Add(10)
		totalV2.WithLabelValues("/root/").Inc()
		totalV2.WithLabelValues("/login/").Add(5)
		cpuPercent.WithLabelValues("0").Set(rand.Float64())
		cpuPercent.WithLabelValues("1").Set(rand.Float64())
		requestTime.WithLabelValues("/root/").Observe(rand.ExpFloat64() * 20)
		requestTime.WithLabelValues("/login/").Observe(rand.ExpFloat64() * 20)
		requestTimeS.WithLabelValues("/root/").Observe(rand.ExpFloat64() * 20)
		requestTimeS.WithLabelValues("/login/").Observe(rand.ExpFloat64() * 20)
	}()

	// 采样值什么时候更新？
	// 1、定时更新
	// 2、metrics api请求时更新 => 采集数据时间会影响api请求时间
	// 3、请求发送时更新，事件触发

	// 暴露http api
	http.Handle("/metrics", promhttp.Handler())
	// 启动web服务
	http.ListenAndServe(addr, nil)
}
