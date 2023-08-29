FROM golang:1.21 AS builder

WORKDIR /go/src/github.com/Octops/agones-relay-http

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN make build && chmod +x /go/src/github.com/Octops/agones-relay-http/bin/agones-relay-http

FROM gcr.io/distroless/static:nonroot

WORKDIR /app

COPY --from=builder /go/src/github.com/Octops/agones-relay-http/bin/agones-relay-http /app/

ENTRYPOINT ["./agones-relay-http"]
