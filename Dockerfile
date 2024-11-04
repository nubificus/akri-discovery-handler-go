FROM docker.io/library/golang:1.22.2 AS builder
WORKDIR /app
COPY . .
RUN go mod tidy && go build

FROM docker.io/library/ubuntu:latest
COPY --from=builder /app/akri-discovery-handler-go /akri-discovery-handler-go
ENTRYPOINT ["/akri-discovery-handler-go"]