# Jaeger  Prometheus Exporter  (Elasticsearch)

> 从opentracing jaeger后端存储Elasticsearch获取并分析数据，统计operation及service的频率，暴露数据给prometheus。

### build
```$xslt
$ cd src/kakko/

$ ls
kakko.go      kakko_test.go

$ GOOS=linux GOARCH=amd64 go build 
$ ls 
kakko         kakko.go      kakko_test.go

$ ./kakko 

```

### 构建镜像
```$xslt
docker build -t reg-tag.xxxx.com/k8s/jaeger_exporter:v1.0 . 

docker run -p 2333:2333 reg-tag.xxx.com/k8s/jaeger_exporter:v1.0
```
### TODO
```$xslt
1. Prometheus query
2. ...
```

### Jaeger Docker-compose
```$xslt
version: '2'
services:
    jaeger-collector:
      image: jaegertracing/jaeger-collector
      ports:
        - "14269"
        - "14268:14268"
        - "14267"
        - "14250"
        - "9411:9411"
      environment:
        - SPAN_STORAGE_TYPE=elasticsearch
        - COLLECTOR_ZIPKIN_HTTP_PORT=9441
        - ES_SERVER_URLS=http://172.23.4.154:32104
        - ES_TAGS_AS_FIELDS=true 

    jaeger-query:
      image: jaegertracing/jaeger-query
      #command: ["--jaeger-agent.host=jaeger-agent"]
      ports:
        - "16686:16686"
        - "16687"
      environment:
        - SPAN_STORAGE_TYPE=elasticsearch
        - ES_SERVER_URLS=http://172.23.4.154:32104
        - ES_TAGS_AS_FIELDS=true

    jaeger-agent:
      image: jaegertracing/jaeger-agent
      command: ["--reporter.type=grpc", "--reporter.grpc.host-port=jaeger-collector:14250"]
      ports:
        - "5775:5775/udp"
        - "6831:6831/udp"
        - "6832:6832/udp"
        - "5778:5778"
```


