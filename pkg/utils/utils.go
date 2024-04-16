package utils

import (
	"context"
	"fmt"

	"github.com/nubificus/akri-discovery-handler-go/pkg/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func RegisterDiscoveryHandler(agentSocket string, discoveryHandlerName string, discoverySocket string) error {
	creds := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.Dial(fmt.Sprintf("unix://%s", agentSocket), creds)
	if err != nil {
		return err
	}
	defer conn.Close()
	client := pb.NewRegistrationClient(conn)
	_, err = client.RegisterDiscoveryHandler(
		context.Background(),
		&pb.RegisterDiscoveryHandlerRequest{
			Name:         discoveryHandlerName,
			Endpoint:     discoverySocket,
			Shared:       true, // Replace with appropriate shared value
			EndpointType: pb.RegisterDiscoveryHandlerRequest_UDS,
		})
	if err != nil {
		return err
	}
	return nil
}
