# akri-discovery-handler-go

A simple implementation of an Akri Discovery Handler written in Go

## How to build akri-discovery-handler-go

A [Dockerfile](./Dockerfile) is provided to facilitate the building process of the container image required to run `akri-discovery-handler-go`:

```bash
docker build -t akri-discovery-handler-go:latest .
docker build --push -t gntouts/akri-discovery-handler-go:scenario2 .
```

You can then push the image to any container registry.

## How to deploy akri-discovery-handler-go

An example configuration `.yaml` that uses this Discovery Handler can be found at [./deploy/exampleConfiguration.yaml](deploy/exampleConfiguration.yaml).

Please note that the [discoveryHandlerName](deploy/exampleConfiguration.yaml#L15) (eg `go-http-rangeswitch`) must be of the format `go-http-range$SUFFIX`, where `$SUFFIX` is defined by [custom.discovery.suffix](deploy/exampleConfiguration.yaml#L37).

To deploy in a existing Akri installation:

```bash
REPO_NAME="nbfc-akri-charts" # edit this accordingly
helm template akri $REPO_NAME/akri -f ./deploy/exampleConfiguration.yaml > template.yaml
kubectl apply -f ./template.yaml
```

## How akri-discovery-handler-go works

The execution flow of any Akri Discovery Handler is fairly simple:

- It initializes any required values, reading from ENV variables.
- It registers itself as an available Discovery Handler to the Akri agent using the provided socket (by default `/var/lib/akri/agent-registration.sock`).
- It creates a new socket (by default `/var/lib/akri/go-http-range$SUFFIX.sock`).
- Listens for any incoming `DiscoverRequest`, parses the DiscoveryDetails given, uses them to discover devices, and responds with a list of the discovered HTTP devices.

## How to actually provide Device Discovery functionality

Currently, this Discovery Handler doesn't discover any devices. In fact, it always responds with the same list containing only one dummy device.
In order to provide an actual discovery service, we need to implement [func (s *server) Discover()](./service.go#L13) to parse the DiscoveryDetails and discover any matching devices.

## Appendix

### Environment Variables

(Our custom version of) Akri provides a few environment variables that are neccessary for this Discovery Handler to function:

| ENV VAR                       | Default Akri  | Extended Akri  |
|------------------------------ |-------------- |--------------- |
| DISCOVERY_HANDLERS_DIRECTORY  | Yes           | Yes            |
| DISCOVERY_HANDLER_SUFFIX      | No            | Yes            |

### How to generate the protobuf pkg

To generate the neccessary package that implements the DiscoveryHandler service and Registration client defined in the [Akri's discovery gRPC proto file](https://github.com/project-akri/akri/blob/main/discovery-utils/proto/discovery.proto):

```bash
curl -L -o discovery.proto https://raw.githubusercontent.com/project-akri/akri/main/discovery-utils/proto/discovery.proto
sed -i '/^package v0;/a option go_package = "github.com/nubificus/akri-discovery-handler-go/pkg/pb";' discovery.proto
mkdir -p pkg/pb
protoc --go_out=./pkg/pb --go_opt=paths=source_relative --go-grpc_out=./pkg/pb --go-grpc_opt=paths=source_relative discovery.proto
```
