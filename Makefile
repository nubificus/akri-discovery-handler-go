COMMIT := $(shell git describe --dirty --long --always)

image:
	docker buildx build --platform linux/arm64,linux/amd64 --push -t harbor.nbfc.io/nubificus/iot/akri-discovery-handler-go:$(COMMIT) -f Dockerfile .
 	