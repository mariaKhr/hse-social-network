FROM golang:1.20-alpine

WORKDIR /main-service
COPY go.* .

RUN go mod download -x

COPY * .

RUN go build .

ENTRYPOINT ["./main-service"]