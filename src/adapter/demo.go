package adapter

import (
    "github.com/prometheus/client_golang/prometheus"
    "sync"
)

// Demo --- Begin

// 指标结构体
type Metrics struct {
    metrics map[string]*prometheus.Desc
    mutex   sync.Mutex
}

/**
 * 函数：newGlobalMetric
 * 功能：创建指标描述符
 */
func newGlobalMetric(metricName string, docString string, labels []string) *prometheus.Desc {
    return prometheus.NewDesc(metricName, docString, labels, nil)
}

/**
 * 工厂方法：NewMetrics
 * 功能：初始化指标信息，即Metrics结构体
 */
func NewMetrics() *Metrics {
    return &Metrics{
        metrics: map[string]*prometheus.Desc{
            "jaeger_operations_duration_6m_requests":      newGlobalMetric("jaeger_operations_duration_6m_requests", "jaeger_operations_duration_6m_requests", []string{"host"}),
            "jaeger_operations_duration_seconds_requests": newGlobalMetric("jaeger_operations_duration_seconds_requests", "The description of jaeger_operations_duration_seconds_requests", []string{"host"}),
        },
    }
}

/**
 * 接口：Describe
 * 功能：传递结构体中的指标描述符到channel
 */
func (c *Metrics) Describe(ch chan<- *prometheus.Desc) {
    for _, m := range c.metrics {
        ch <- m
    }
}

/**
 * 接口：Collect
 * 功能：抓取最新的数据，传递给channel
 */
func (c *Metrics) Collect(ch chan<- prometheus.Metric) {
    c.mutex.Lock() // 加锁
    defer c.mutex.Unlock()

    mockCounterMetricData, mockGaugeMetricData := c.GenerateMockData()

    for host, currentValue := range mockCounterMetricData {
        // 关键是这个 这个数据怎么封装 第一个参数*Desc,第二个数据类型，第三个value，再往后lableValues ...string
        ch <- prometheus.MustNewConstMetric(c.metrics["jaeger_operations_duration_6m_requests"], prometheus.GaugeValue, float64(currentValue), host)
    }
    for host, currentValue := range mockGaugeMetricData {
        ch <- prometheus.MustNewConstMetric(c.metrics["jaeger_operations_duration_seconds_requests"], prometheus.GaugeValue, float64(currentValue), host)
    }
}

/**
 * 函数：GenerateMockData
 * 功能：生成模拟数据
 */
func (c *Metrics) GenerateMockData() (mockCounterMetricData map[string]int, mockGaugeMetricData map[string]int) {
    mockCounterMetricData = map[string]int{
        "ertyuiop": 100,
    }
    mockGaugeMetricData = map[string]int{
        "zxcvbnvbc": 102,
    }
    return
}

// Demo --- End