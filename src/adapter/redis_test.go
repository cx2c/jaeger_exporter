package adapter

import (
    "fmt"
    "testing"
    "github.com/garyburd/redigo/redis"
    "reflect"
    "encoding/json"
    "time"
)

var NUM int64

type errorInfo struct {
    StartTimeMillis int64
    operationNmae   string
    service         string
    errorMessage    string
    spanID          string
}

var mxMapList []map[string]errorInfo

func TestRedisCurd(t *testing.T) {
    c, err := redis.Dial("tcp", "127.0.0.1:6379")
    if err != nil {
        fmt.Println("Connect redis error !", err)
        return
    }
    defer c.Close()

    // keys *
    c.Do("select", "3")
    n, err := c.Do("keys", "*", )
    if err != nil {
        fmt.Println("keys * error", err)
    } else {
        fmt.Println(reflect.TypeOf(n))
        //fmt.Println(reflect.ValueOf(n))
        fmt.Printf("%s \n", n)
    }

    // keys * 获取字符串
    re, err := redis.String(c.Do("get", "addr"))
    if err != nil {
        fmt.Println("kesy * string error", err)
    } else {
        fmt.Println(re)
        fmt.Println(reflect.TypeOf(re))
    }

    // 写入数据  - 怎么处理类似于dict类型：用map来解决
    // 思路就是 marshal & unmashall ，根据，map定义的格式进行解码
    mMap := map[string]string{"operationName": "getToken", "service": "aipower", "message": "rtyu"}

    // map 转json, 转成bytes二进制类型
    value, _ := json.Marshal(mMap)
    fmt.Println("bytes 类型？：", string(value))

    ns, err := c.Do("set", "abc", value)

    if err != nil {
        fmt.Println("error when set :", err)
    } else {
        fmt.Println(ns)
    }
    // 验证
    nw, err := redis.Bytes(c.Do("get", "abc"))
    if err != nil {
        fmt.Println("error when set :", err)
    } else {
        fmt.Println("123:", reflect.TypeOf(nw))
        err := json.Unmarshal(nw, &mMap)
        fmt.Println("123:", reflect.TypeOf(nw))
        if err != nil {
            fmt.Println(err)
        } else {
            fmt.Println("marshal 之后的nw：", mMap["operationName"])
        }
    }

    // []map[int64]errorInfo{} 这种数据结构，key 就是 spanID
    // 一个读写操作 map[string]errorInfo
    var ssMap = map[string]errorInfo{"ads": {11111, "dsadas", "ddd", "dddd", "ffff"}}
    var ss2Map = map[string]errorInfo{"ads": {11111, "dsadas", "ddd", "dddd", "dddd"}}
    //mxMapList[0] = ssMap

    // 取service
    fmt.Println(ssMap["ads"].service)
    fmt.Println(reflect.TypeOf(ssMap["ads"]))

    mxMapList = append(mxMapList, ssMap)
    mxMapList = append(mxMapList, ss2Map)
    fmt.Println("mxMapList:", mxMapList)
    fmt.Println("mxMapList:", len(mxMapList))

    // 写入redis
    value3, _ := json.Marshal(mxMapList)
    fmt.Println("map[string]errorInfo类型：", string(value3))
    nsw, errw := c.Do("set", "lianbiao", value3)
    if errw != nil {
        fmt.Println("error", errw)
    } else {
        fmt.Println(nsw)
    }

    // 生成timestrap 毫秒
    if NUM == 0 {
        NUM = int64(time.Now().Unix()) * 1000
    }
    //for {
    //    go func() {
    //        // 这里判断，如果NUM值不存在 需要初始
    //
    //        NUM += 10000
    //        fmt.Println("running")
    //
    //    }()
    //    fmt.Println(NUM)
    //    time.Sleep(5 * time.Second)
    //}
}
