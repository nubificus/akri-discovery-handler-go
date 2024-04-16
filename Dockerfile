FROM golang:1.22.2 as builder
WORKDIR /app
COPY . .
RUN go mod tidy && go build

FROM ubuntu
COPY --from=builder /app/akri-discovery-handler-go /akri-discovery-handler-go
ENTRYPOINT ["/akri-discovery-handler-go"]