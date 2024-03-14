FROM golang:1.20-alpine

# RUN apt update
# RUN apt install -y postgresql
# RUN systemctl start postgresql.service

WORKDIR /main-service
COPY go.* .

RUN go mod download -x

COPY * .

RUN go build .

ENTRYPOINT ["./main-service"]