FROM golang:1.10


ENV DEBUG 1
ENV PROFILE_PORT 6060
ENV STATSD_HOST dogstatsd:8125

WORKDIR /go/src/github.com/embrace-io/dbr

CMD ["bash", "-c", "sh run.sh"]
