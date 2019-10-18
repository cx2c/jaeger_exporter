# Prometheus Jaeger  Exporter  (Elasticsearch)
 
> Opentracing Jaeger. Form Elasticsearch obtain and analyze datacalculate the frequency ofoperation and service, And expose 
> the data to Prometheus. Analyzes and gets opentracing error from Elasticsearch Prometheus completed according to the error  > alarm is triggered

### dependent
```$xslt
go get github.com/elastic/go-elasticsearch
go get gopkg.in/olivere/elastic.v6
go get github.com/prometheus/client_golang
go get github.com/prometheus/client_golang/prometheus
go get github.com/garyburd/redigo/redis
```
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

### docker
```$xslt
docker build -t reg-tag.xxxx.com/k8s/jaeger_exporter:v1.0 . 

docker run -p 2333:2333 reg-tag.xxx.com/k8s/jaeger_exporter:v1.0
```
### TODO
```$xslt
1.
2.
```

### Jaeger Docker-compose
```$xslt
version: '2'
services:
  # els:
  #   image: docker.elastic.co/elasticsearch/elasticsearch:6.0.0
  #   restart: always
  #   container_name: els
  #   hostname: els
  #   networks:
  #   - elastic-jaeger
  #   environment:
  #     - bootstrap.memory_lock=true
  #     - ES_JAVA_OPTS=-Xms512m -Xmx512m
  #   ports:
  #     - "9200:9200"
  #   ulimits:
  #     memlock:
  #       soft: -1
  #       hard: -1
  #   mem_limit: 1g
  #   volumes:
  #     - esdata1:/usr/share/elasticsearch/data
  #     - eslog:/usr/share/elasticsearch/logs
  #     - ./config/elasticsearch.yml:/usr/share/elasticsearch/config/elasticsearch.yml
  # kibana:
  #   image: docker.elastic.co/kibana/kibana:6.0.0
  #   ports:
  #     - "5601:5601"
  #   environment:
  #     ELASTICSEARCH_URL: http://els:9200
  #   depends_on:
  #   - els
  #   networks:
  #   - elastic-jaeger
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

# Jaeger Spark dependencies

### run with docker and Kubernetes
```$xslt
kubectl run --restart=Never spark-dependencies --env="STORAGE=elasticsearch" --env="ES_NODES=http://elasticsearch-logging:9200" --image="jaegertracing/spark-dependencies" -n kube-system
 docker run  --rm  --name  spark-dependencies  --env STORAGE=elasticsearch --env ES_NODES=http://172.23.4.154:32104/ jaegertracing/spark-dependencies
```


### spark-dependencies cronjab In Kubernetes
```$xslt
apiVersion: batch/v1beta1
kind: CronJob
metadata:
  namespace: kube-system
  name: spark-dependencies
  labels:
    app.kubernetes.io/name: spark-dependencies
spec:
  schedule: "*/10 * * * *"
  concurrencyPolicy: Forbid
  jobTemplate:
    spec:
      template:
        spec:
          containers:
            - name: spark-dependencies
              image: jaegertracing/spark-dependencies
              imagePullPolicy: IfNotPresent
              env:
              - name: STORAGE
                value: elasticsearch
                name: ES_NODES
                value: http://elasticsearch-logging:9200
          restartPolicy: OnFailure
```


# In Kubernetes
```
.kubernetes/jaeger_exporter.yaml        ## yaml of prometheus exporter in Kubernetes 
.kubernetes/jaeger_monitoring.yaml      ## add prometheus target for this exporter
.kubernetes/Jaeger-grafana.json         ## grafana dashboard 
.kubernetes/jaeger_dependencies.yaml    ## jaeger dependencies spark task in Kubernetes cronjob
.kubernetes/es_hq.yaml                  ## es management tool ElasticsrarchHQ in Kubernetes
``` 
