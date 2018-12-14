# Jaeger es 数据计算 demo

> jaeger prometheus exporter  
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
1.
2.
```
