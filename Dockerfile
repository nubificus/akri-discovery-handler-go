FROM harbor.nbfc.io/proxy_cache/library/golang:1.24.2-alpine3.21 AS builder
WORKDIR /app
COPY . .
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -ldflags "-extldflags '-static'"

FROM harbor.nbfc.io/proxy_cache/library/alpine:3.21

COPY --from=builder /app/akri-discovery-handler-go /akri-discovery-handler-go
ENTRYPOINT ["/akri-discovery-handler-go"]