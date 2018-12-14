package main

import (
    "gopkg.in/olivere/elastic.v6"
    "log"
    "os"
    "context"
    "fmt"
    "time"
    "strings"
    "reflect"
    "encoding/json"
    "github.com/prometheus/client_golang/prometheus"
    "net/http"
    "runtime"
    "strconv"
    "sync"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

// 取gorountine id 暂时不用，作者说 用此方法便入地狱 🙄️
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

// 初始化es客户端 注意会解析k8s里es集群的真实地址 ，http://10.1.4.46:9200 ,放在k8s里跑也不会有影响
var client *elastic.Client
var host = "http://elasticsearch-svc:9200"
//var host = "http://192.168.30.240:9200"
//var host = "http://172.23.4.154:32104/"


// 数据结构 测试 TraceGroup SpanInfo Process SpanSource
type TraceGroup struct {
    RefType string `json:"ref_type"`
    TraceID string `json:"trace_id"`
    SpanId  string `json:"span_id"`
}

type SpanInfo struct {
    TraceID string `json:"trace_id"`
    SpanId  string `json:"span_id"`
}

type Process struct {
    ServiceName string     `json:"service_name"`
    Tags        [10]string `json:"tags"`
}

type SpanSource struct {
    SpanInfo        SpanInfo
    StartTime       int64      `json:"starttime"`         // 微秒
    StartTimeMillis int64      `json:"start_time_millis"` // 毫秒
    Duration        int64      `json:"duration_time"`     // 微秒
    Flags           string     `json:"flags"`
    OperationName   string     `json:"operation_name"`
    References      TraceGroup
    Tags            [10]string `json:"tags"`
    Process         Process
}

// 采用此struct 匹配es里面数据
type Tmp struct {
    TraceID         string                   `json:"traceID"`
    SpanID          string                   `json:"spanID"`
    StartTime       int64                    `json:"startTime"`       // 微秒
    StartTimeMillis int64                    `json:"startTimeMillis"` // 毫秒
    Duration        int64                    `json:"duration_time"`   // 微秒
    Flags           int                      `json:"flags"`
    OperationName   string                   `json:"operationName"`
    References      []string                 `json:"references,omitempty"`
    Tags            []map[string]interface{} `json:"tags"` //奇怪 reference 能取到，但是tags 和process 取不到
    Process         map[string]interface{}   `json:"process,omitempty"`
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

// Demo Begin
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

// Demo End

func init() {
    errorlog := log.New(os.Stdout, "Jaeger ", log.LstdFlags)
    var err error
    client, err = elastic.NewClient(elastic.SetErrorLog(errorlog), elastic.SetURL(host))
    if err != nil {
        panic(err)
    }
    info, code, err := client.Ping(host).Do(context.Background())
    if err != nil {
        panic(err)
    }
    fmt.Printf("Elasticsearch returned with code %d and version %s \n", code, info.Version.Number)

    esversion, err := client.ElasticsearchVersion(host)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Elasticsearch version %s\n", esversion)

}

// 获取idnex 索引头
func getIndexname() string {
    indexHead := "jaeger-span-"
    now := time.Now().String()
    nowStr := strings.Split(now, " ")
    nowYmd := nowStr[0]
    indexName := indexHead + nowYmd
    return indexName
}

// 测试方法 golang 操作es
func Curd() {
    indexName := getIndexname()

    // get 查询  --必须跟id
    get1, err := client.Get().Index(indexName).Type("span").Id("YlQpm2cBSNOZuX2mbAQC").Do(context.Background())
    if err != nil {
        panic(err)
    }
    if get1.Found {
        fmt.Printf("Got document %s in version %d from index %s, type %s\n", get1.Id, get1.Version, get1.Index, get1.Type)
        fmt.Println("source: ", string(*get1.Source))
    }
    /***   Result:
    Got document tixplWcBpolvgvsA1x0S in version 842350569472 from index jaeger-span-2018-12-10, type span
    source:  {"traceID":"84651b191a13fd1","spanID":"5732c5e0e23e1cea","flags":1,"operationName":"Execute",
              "references":[{"refType":"CHILD_OF","traceID":"84651b191a13fd1","spanID":"84651b191a13fd1"}],
              "startTime":1544400003034000,
                "startTimeMillis":1544400003034,
                "duration":14185,
                "tags":[{"key":"component","type":"string","value":"java-jdbc"},{"key":"db.type","type":"string","value":"mysql"},
                        {"key":"db.user","type":"string","value":"ntreader@172.23.0.39"},{"key":"span.kind","type":"string","value":"client"},{"key":"db.statement","type":"string","value":"INSERT INTO crm_customer_source  ( source_id,site_id,type_id,type_name,pid,create_date ) VALUES( ?,?,?,?,?,? )"}],"logs":[],"process":{"serviceName":"CRM","tags":[{"key":"hostname","type":"string","value":"Dora-PC"},{"key":"jaeger.version","type":"string","value":"Java-0.31.0"},{"key":"ip","type":"string","value":"192.168.30.93"}]}}
    ***/

    // search 检索
    // termQuery 理解成单词查询  根据operationName来筛选
    fmt.Println("begine... ...")
    ctx := context.Background()
    termQuery := elastic.NewTermQuery("operationName", "/api/services")
    searchResult, err := client.Search().
        Index(indexName).
        Query(termQuery).
        Sort("operationName", true).From(0).Size(1000).Pretty(true).Do(ctx)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Query took %d milliseconds\n", searchResult.TookInMillis)
    fmt.Println(*searchResult.Hits)

    // 打印 Hits.Hits xFTXm2cBSNOZuX2mah3z jaeger-span-2018-12-11 span json: Unmarshal(non-pointer kakko.SpanSource)
    //for _, i :=range searchResult.Hits.Hits{
    //    fmt.Println(string(i.Id),string(i.Index),string(i.Type),json.Unmarshal(*i.Source,SpanSource{}))
    //}

    var ttype Tmp
    for _, item := range searchResult.Each(reflect.TypeOf(ttype)) {
        t := item.(Tmp)
        fmt.Println(t.TraceID, t.SpanID, t.Flags, t.OperationName, t.StartTime, t.References, t.Duration, t.Process["serviceName"], t.Process["tags"], ":-1-1-1:", t.Tags)
        //for _, i :=range t.Tags{
        //    for k,v := range i{
        //        fmt.Println(k,v)
        //    }
        //}
    }
    fmt.Println("done... ...")
    fmt.Println("----------")
    fmt.Println("----------")
    fmt.Println("----------")
    fmt.Println("----------")
    // 根据 时间来筛选
    fmt.Println("begine2... ...")
    rangeQuery := elastic.NewRangeQuery("startTimeMillis").Gte(1542499480006).Lte(1544499480006)
    src, err := rangeQuery.Source()
    data, err := json.Marshal(src)
    got := string(data)
    fmt.Println("range 得到的query？？？？？：", got)

    searchResult2, err := client.Search().
        Index(indexName).
        Query(rangeQuery).From(0).Size(1000).Pretty(true).Do(ctx)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Query took %d milliseconds\n", searchResult2.TookInMillis)
    var ttype2 Tmp
    for _, item := range searchResult2.Each(reflect.TypeOf(ttype2)) {
        t := item.(Tmp)
        fmt.Println(t.TraceID, t.SpanID, t.Flags, t.OperationName, t.StartTime, t.References, t.Duration, t.Process["serviceName"], t.Process["tags"], ":-1-1-1:", t.Tags)
        for _, i := range t.Tags {
            for k, v := range i {
                fmt.Println(k, v)
            }
        }
    }
    fmt.Println("done2... ...")
}

//从es里取筛选数据-from to 一段时间
func GetElasticsearch(seconds int64) *elastic.SearchResult {
    ctx := context.Background()
    indexName := getIndexname()
    timeNowMill := int64(time.Now().UnixNano() / 1e6)
    timeNowMillB5m := timeNowMill - 1000*seconds
    rangeQuery := elastic.NewRangeQuery("startTimeMillis").Gte(timeNowMillB5m).Lte(timeNowMill)
    src, err := rangeQuery.Source()
    data, err := json.Marshal(src)
    got := string(data)
    fmt.Println("range 得到的query：", got)

    searchResult, err := client.Search().
        Index(indexName).
        Query(rangeQuery).From(0).Size(1000).Pretty(true).Do(ctx)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Query took %d milliseconds\n", searchResult.TookInMillis)
    //var ttype2 Tmp
    //for _, item := range searchResult.Each(reflect.TypeOf(ttype2)) {
    //   t := item.(Tmp)
    //   OperationName := t.OperationName
    //   fmt.Println("5M函数打印：",OperationName)
    //   fmt.Println("5M函数打印：",t.TraceID, t.SpanID, t.Flags, t.OperationName, t.StartTime, t.References, t.Duration, t.Process["serviceName"], t.Process["tags"], ":-1-1-1:", t.Tags)
    //   //for _, i := range t.Tags {
    //   //    for k, v := range i {
    //   //        fmt.Println(k, v)
    //   //    }
    //   }

    return searchResult

}

// prometheus init
func Prom() (*prometheus.GaugeVec, *prometheus.GaugeVec, *prometheus.Registry) {
    jaegerDuration5MRequests := promauto.NewGaugeVec(prometheus.GaugeOpts{
        Name: "jaeger_operations_duration_5m_requests",
        Help: "different operations in 5m ",
    }, []string{"operationname", "service"})

    jaegerQPSRequests := promauto.NewGaugeVec(prometheus.GaugeOpts{
        Name: "jaeger_operations_qps_requests",
        Help: "different operations in 1m  NOT 1s! ",
    }, []string{"operationname", "service"})

    registry := prometheus.NewRegistry()

    registry.MustRegister(jaegerDuration5MRequests)
    registry.MustRegister(jaegerQPSRequests)

    return jaegerDuration5MRequests, jaegerQPSRequests, registry
}

func runData() {
    go func() {

        runtime.GOMAXPROCS(4)
        var wg sync.WaitGroup

        // origin
        jaegerDuration5MRequests, jaegerQPSRequests, registry := Prom()

        // 下面两种方法都不生效 ，得到的inc() 还是不断累加的
        // way1
        //registry.Unregister(jaegerDuration5MRequests)
        //registry.Unregister(jaegerQPSRequests)
        //registry.MustRegister(jaegerDuration5MRequests)
        //registry.MustRegister(jaegerQPSRequests)

        // way2
        //jaegerDuration5MRequests.Reset()
        //jaegerQPSRequests.Reset()

        // 取es数据
        searchResult5M := GetElasticsearch(300 )
        searchResult1S := GetElasticsearch(60 )
        fmt.Println("searchResult:", *searchResult5M)

        // ♻️♻️♻️♻️
        for {

            // 每次循环之前 先把labelsArray 里的metrics+labels 都删掉
            // 再把 labelsArray 数据清掉
            // 清空所有label

            //取计数器值 遍历 labelsArray5M 并删除metrics
            for _, labels := range labelsArray5M {
                operationname := labels.operationNmae
                service := labels.service
                jaegerDuration5MRequests.Delete(prometheus.Labels{"operationname": operationname, "service": service})
            }
            //取计数器值 遍历 labelsArray1S 并删除metrics
            for _, labels := range labelsArray1S {
                operationname := labels.operationNmae
                service := labels.service
                jaegerQPSRequests.Delete(prometheus.Labels{"operationname": operationname, "service": service})
            }

            // 初始化计数器
            counter5M = 0
            counter1S = 0

            // 加锁
            var ttype Tmp
            var lock sync.Mutex
            var lock2 sync.Mutex

            // 计算五分钟==300秒的请求次数
            for _, item := range searchResult5M.Each(reflect.TypeOf(ttype)) {
                wg.Add(1)
                go func() {
                    defer wg.Done()
                    t := item.(Tmp)
                    str := t.Process["serviceName"].(string)
                    OperationName := t.OperationName
                    lock.Lock()
                    jaegerDuration5MRequests.With(prometheus.Labels{"operationname": OperationName, "service": str}).Inc()
                    lock.Unlock()
                    labelsArray5M[counter5M] = labelsMap5M{OperationName, str}
                    counter5M += 1
                }()
                wg.Wait()
            }
            // 计算1秒==1秒的请求次数
            for _, item := range searchResult1S.Each(reflect.TypeOf(ttype)) {
                wg.Add(1)
                go func() {
                    defer wg.Done()
                    t := item.(Tmp)
                    str := t.Process["serviceName"].(string)
                    OperationName := t.OperationName
                    lock2.Lock()
                    jaegerQPSRequests.With(prometheus.Labels{"operationname": OperationName, "service": str}).Inc()
                    lock2.Unlock()
                    labelsArray1S[counter5M] = labelsMap1S{OperationName, str}
                    counter1S += 1
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

        metrics2 := NewMetrics()
        registry.MustRegister(metrics2)

    }()

}
func main() {

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

// 下面是重写 promhttp.Handler()

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
