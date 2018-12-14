FROM        quay.io/prometheus/busybox:latest
MAINTAINER  zhangwei  <zhangwei@xiaoneng.cn>

COPY src/kakko/kakko /bin/kakko

EXPOSE      2333
ENTRYPOINT  [ "/bin/kakko" ]