FROM golang:1.20-alpine

ADD . /go/src/main-service

WORKDIR /go/src/main-service

RUN go get main-service && go install

ENTRYPOINT ["/go/bin/main-service"]