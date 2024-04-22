FROM golang:1.20-alpine

ADD . /go/src/post-service

WORKDIR /go/src/post-service

RUN go get post-service && go install

ENTRYPOINT ["/go/bin/post-service"]