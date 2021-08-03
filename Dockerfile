FROM quay.io/prometheus/golang-builder AS builder

ADD .   /go/src/github.com/nelsonjohnstone/cmc_exporter
WORKDIR /go/src/github.com/nelsonjohnstone/cmc_exporter

RUN make

FROM        quay.io/prometheus/busybox:glibc
LABEL maintainer="The Prometheus Authors <prometheus-developers@googlegroups.com>"
COPY        --from=builder /go/src/github.com/nelsonjohnstone/cmc_exporter/cmc_exporter  /bin/cmc_exporter

EXPOSE      9599
ENTRYPOINT  [ "/bin/cmc_exporter" ]