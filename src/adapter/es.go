package adapter

import (
    "gopkg.in/olivere/elastic.v6"
    "log"
    "os"
    "context"
    "fmt"
    "reflect"
    "encoding/json"
    "time"
    "strings"
)

// 初始化es客户端 注意会解析k8s里es集群的真实地址 ，http://10.1.4.46:9200 ,放在k8s里跑也不会有影响
var client *elastic.Client
var host = "http://elasticsearch-logging.kube-system:9200"
//var host = "http://192.168.30.240:9200"
//var host = "http://172.23.4.154:32104/"

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
func GetElasticsearch(seconds int64, tag int) (*elastic.SearchResult, int64) {
    // tag = 1 , seconds是秒，取多少秒的数据
    // tag = 2 ， seconds是timestrap，从什么时间开始取
    ctx := context.Background()
    indexName := getIndexname()
    timeNowMill := int64(time.Now().UnixNano() / 1e6)
    var timeNowMillB5m int64

    if tag == 2 {
        timeNowMillB5m = seconds
        fmt.Println("get elasticsearch data into mode 2")
    } else {
        timeNowMillB5m = timeNowMill - 1000*seconds
    }
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
    return searchResult, timeNowMill

}
