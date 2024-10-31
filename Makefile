image:
	docker buildx build --platform linux/arm64,linux/amd64 --push -t harbor.nbfc.io/nubificus/iot/akri-discovery-handler-go:latest -f Dockerfile .