package main

import (
    "adapter"
    _ "adapter"
    "log"
    "fmt"
    "time"
    "strings"
    "reflect"
    "github.com/prometheus/client_golang/prometheus"
    "net/http"
    "runtime"
    "strconv"
    "sync"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

// 取gorountine id 暂时不用，作者说 用此方法便入地狱😓
func GoIDs() int {
    var buf [64]byte
    n := runtime.Stack(buf[:], false)
    idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
    id, err := strconv.Atoi(idField)
    if err != nil {
        panic(fmt.Sprintf("cannot get goroutine id: %v", err))
    }
    return id
}

// 清除metrics labels使用
type labelsMap5M struct {
    operationNmae string
    service       string
}

var (
    labelsArray5M [2000]labelsMap5M
    counter5M     int
)

type labelsMap1S struct {
    operationNmae string
    service       string
}

var (
    labelsArray1S [2000]labelsMap1S
    counter1S     int
)

type labelsMap1MError struct {
    operationNmae string
    service       string
}

var (
    labelsArray1MError [2000]labelsMap1MError
    counter1MError     int
    LASTTIME           int64
)

// prometheus init
func Prom() (*prometheus.GaugeVec, *prometheus.GaugeVec, *prometheus.GaugeVec, *prometheus.Registry) {
    jaegerDuration5MRequests := promauto.NewGaugeVec(prometheus.GaugeOpts{
        Name: "jaeger_operations_duration_5m_requests",
        Help: "different operations in 5m ",
    }, []string{"operationname", "service"})

    jaegerQPSRequests := promauto.NewGaugeVec(prometheus.GaugeOpts{
        Name: "jaeger_operations_qps_requests",
        Help: "different operations in 1m  NOT 1s! ",
    }, []string{"operationname", "service"})

    jaegerDuration1MErrors := promauto.NewGaugeVec(prometheus.GaugeOpts{
        Name: "jaeger_operations_duration_1m_errors",
        Help: "errs of spans in last 1m ",
    }, []string{"operationname", "service"})

    registry := prometheus.NewRegistry()

    registry.MustRegister(jaegerDuration5MRequests)
    registry.MustRegister(jaegerQPSRequests)
    registry.MustRegister(jaegerDuration1MErrors)

    return jaegerDuration5MRequests, jaegerQPSRequests, jaegerDuration1MErrors, registry
}

// 循环 生成mtrices 及数据
func runData() {
    go func() {

        runtime.GOMAXPROCS(4)
        var wg sync.WaitGroup

        // origin
        jaegerDuration5MRequests, jaegerQPSRequests, jaegerDuration1MErrors, registry := Prom()

        // ♻️♻️♻️♻️
        for {
            fmt.Println("LASTTIME:", LASTTIME)
            // 取es数据
            searchResult5M, _ := adapter.GetElasticsearch(300, 1)
            searchResult1S, _ := adapter.GetElasticsearch(60, 1)
            searchResult1MError, timeNowMill := adapter.GetElasticsearch(LASTTIME, 2)
            LASTTIME = timeNowMill
            fmt.Println("LASTTIME:", LASTTIME)

            // 每次循环之前 先把labelsArray 里的metrics+labels 都删掉
            // 再把 labelsArray 数据清掉
            // 清空所有label

            //取计数器值 遍历 labelsArray5M 并删除metrics
            for _, labels := range labelsArray5M {
                operationname := labels.operationNmae
                service := labels.service
                jaegerDuration5MRequests.Delete(prometheus.Labels{"operationname": operationname, "service": service})
                // 打印日志 生产打开
                //if service != "" {
                //    fmt.Println("删除metrics:", service, operationname)
                //}
            }
            //取计数器值 遍历 labelsArray1S 并删除metrics
            for _, labels := range labelsArray1S {
                operationname := labels.operationNmae
                service := labels.service
                jaegerQPSRequests.Delete(prometheus.Labels{"operationname": operationname, "service": service})
            }
            //取计数器值 遍历 labelsArray1MError 并删除metrics
            for _, labels := range labelsArray1MError {
                operationname := labels.operationNmae
                if labels.service != "" {fmt.Println("删除之前打印：",operationname,1,labels.service)}
                service := labels.service
                jaegerDuration1MErrors.Delete(prometheus.Labels{"operationname": operationname, "service": service})
                //打印日志 生产打开
                if service != "" {
                   fmt.Println("删除metrics222:", service, operationname)
                }
            }

            // 初始化计数器
            counter5M = 0
            counter1S = 0
            counter1MError = 0

            // 加锁
            var ttype adapter.Tmp
            var lock sync.Mutex
            var lock2 sync.Mutex
            var lock3 sync.Mutex

            // 计算五分钟==300秒的请求次数
            for _, item := range searchResult5M.Each(reflect.TypeOf(ttype)) {
                wg.Add(1)
                go func() {
                    defer wg.Done()
                    t := item.(adapter.Tmp)
                    str := t.Process["serviceName"].(string)
                    OperationName := t.OperationName
                    //fmt.Println("生产metrics:", str, OperationName)
                    //fmt.Println("tags:",t.Tags)

                    lock.Lock()
                    jaegerDuration5MRequests.With(prometheus.Labels{"operationname": OperationName, "service": str}).Inc()
                    //fmt.Println("消费metrics:", str, OperationName)
                    lock.Unlock()
                    labelsArray5M[counter5M] = labelsMap5M{OperationName, str}
                    counter5M += 1
                }()
                wg.Wait()
            }

            // 计算1秒==1秒的请求次数     -- add 取错误也从这里取，封装metrics jaeger_operations_duration_1m_errors
            for _, item := range searchResult1S.Each(reflect.TypeOf(ttype)) {
                wg.Add(1)
                go func() {
                    defer wg.Done()
                    t := item.(adapter.Tmp)
                    str := t.Process["serviceName"].(string)
                    OperationName := t.OperationName

                    for _, j := range t.Tags {
                        //fmt.Println("type,key,value: ",j["type"], j["key"], j["value"])
                        //fmt.Println(j["value"])
                        if j["value"] == 500 && j["key"] == "http.status_code" {
                            fmt.Println("此tag 出错 500！")
                        }
                    }

                    //fmt.Println("1S-serviceName-OperationName:", str, OperationName)
                    lock2.Lock()
                    jaegerQPSRequests.With(prometheus.Labels{"operationname": OperationName, "service": str}).Inc()
                    lock2.Unlock()
                    labelsArray1S[counter1S] = labelsMap1S{OperationName, str}
                    counter1S += 1
                }()
                // 等待异步完全执行完
                wg.Wait()
            }

            // 计算过去15秒错误，封装metrics jaeger_operations_duration_1m_errors
            for _, item := range searchResult1MError.Each(reflect.TypeOf(ttype)) {
                wg.Add(1)
                go func() {
                    defer wg.Done()
                    t := item.(adapter.Tmp)
                    str := t.Process["serviceName"].(string)
                    OperationName := t.OperationName
                    for _, j := range t.Tags {
                        //fmt.Println("type,key,value: ",j["type"], j["key"], j["value"])
                        //fmt.Println(j["value"])
                        if j["key"] == "error" {
                            fmt.Println("dfghjk:", j["value"])
                            fmt.Println(reflect.TypeOf(j["value"]))
                            fmt.Printf("\n")
                        }

                        if j["value"] == "true" && j["key"] == "error" {
                            fmt.Println("此 span error:", OperationName)
                            //fmt.Println("1S-serviceName-OperationName:", str, OperationName)
                            lock3.Lock()
                            jaegerDuration1MErrors.With(prometheus.Labels{"operationname": OperationName, "service": str}).Inc()
                            lock3.Unlock()
                            labelsArray1MError[counter1MError] = labelsMap1MError{OperationName, str}
                            counter1MError += 1
                        }
                    }
                }()
                // 等待异步完全执行完
                wg.Wait()
            }

            time.Sleep(15 * time.Second)
        }

        // 测试其他方法
        buildInfo := prometheus.NewGaugeVec(prometheus.GaugeOpts{
            Name: "redis_exporter_build_info",
            Help: "redis exporter build_info",
        }, []string{"version"})
        buildInfo.WithLabelValues("version").Inc()

        registry.Register(buildInfo)

        metrics2 := adapter.NewMetrics()
        registry.MustRegister(metrics2)

    }()

}
func main() {

    // 初始化时间
    if LASTTIME == 0 {
        LASTTIME = int64(time.Now().UTC().UnixNano() / 1e6) + 28800000    // 解决时区问题
    }
    // 刷新数据
    runData()

    // start
    fmt.Println("server is running on http://127.0.0.1:2333/metrics")
    http.Handle("/metrics", promhttp.Handler())

    //旧方法
    //handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
    //http.Handle("/metrics", handler)

    // Home Page
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte(`<html>
             <head><title>Jaeger Exporter</title></head>
             <body>
             <h1>Dummy Exporter</h1>
             <p><a href='` + "/metrics" + `'>Metrics</a></p>
             </body>
             </html>`))
    })
    log.Fatal(http.ListenAndServe(":2333", nil))
}

// 重写 promhttp.Handler()

func Handler() http.Handler {
    return InstrumentMetricHandler(
        prometheus.DefaultRegisterer, promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{}),
    )
}

func InstrumentMetricHandler(reg prometheus.Registerer, handler http.Handler) http.Handler {
    cnt := prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "promhttp_metric_handler_requests_total",
            Help: "Total number of scrapes by HTTP status code.",
        },
        []string{"code"},
    )
    // Initialize the most likely HTTP status codes.
    cnt.WithLabelValues("200")
    cnt.WithLabelValues("500")
    cnt.WithLabelValues("503")
    if err := reg.Register(cnt); err != nil {
        if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
            cnt = are.ExistingCollector.(*prometheus.CounterVec)
        } else {
            panic(err)
        }
    }

    gge := prometheus.NewGauge(prometheus.GaugeOpts{
        Name: "promhttp_metric_handler_requests_in_flight",
        Help: "Current number of scrapes being served.",
    })
    if err := reg.Register(gge); err != nil {
        if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
            gge = are.ExistingCollector.(prometheus.Gauge)
        } else {
            panic(err)
        }
    }

    return promhttp.InstrumentHandlerCounter(cnt, InstrumentHandlerInFlight(gge, handler))
}
func InstrumentHandlerInFlight(g prometheus.Gauge, next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        g.Inc()
        defer g.Dec()
        next.ServeHTTP(w, r)
    })
}
