package adapter

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
    //References      []string                 `json:"references,omitempty"`
    References      []map[string]interface{}                 `json:"references,omitempty"`
    Tags            []map[string]interface{} `json:"tags"` //奇怪 reference 能取到，但是tags 和process 取不到
    Process         map[string]interface{}   `json:"process,omitempty"`
}

// opentracing error
type errorInfo struct {
    StartTimeMillis int64
    operationNmae   string
    service         string
    errorMessage    string
    spanID          string
}

//var ssMap = map[string]errorInfo{"ads": {11111, "dsadas", "ddd", "dddd", "ffff"}}