FROM golang:1.17.0-alpine
WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY *.go ./

RUN go build -o /node-discovery
ENTRYPOINT ["/node-discovery"]

