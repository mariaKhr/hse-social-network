FROM golang:1.21-alpine

ADD . /go/src/statistic-service

WORKDIR /go/src/statistic-service

RUN go get statistic-service && go install

ENTRYPOINT ["/go/bin/statistic-service"]