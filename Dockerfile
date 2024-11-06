FROM docker.io/library/golang:1.22.2 AS builder
WORKDIR /app
COPY . .
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -ldflags "-extldflags '-static'"

FROM harbor.nbfc.io/nubificus/static:latest

COPY --from=builder /app/akri-discovery-handler-go /akri-discovery-handler-go
ENTRYPOINT ["/akri-discovery-handler-go"]