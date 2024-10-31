FROM docker.io/library/golang:1.22.2 AS builder
WORKDIR /app
COPY . .
RUN go mod tidy && go build

FROM gcr.io/distroless/static-debian12:latest
COPY --from=builder /app/akri-discovery-handler-go /akri-discovery-handler-go
ENTRYPOINT ["/akri-discovery-handler-go"]