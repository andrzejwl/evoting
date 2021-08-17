FROM golang:1.17.0-alpine
WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY *.go ./
COPY pbft/*.go ./pbft/
COPY pow/*.go ./pow/

RUN go build -o /evoting
ENTRYPOINT ["/evoting", "-consensus=pbft", "-port=1337"]

