package main

import (
	"fmt"
	"time"

	"github.com/nubificus/akri-discovery-handler-go/pkg/pb"
)

type server struct {
	pb.UnimplementedDiscoveryHandlerServer
}

func (s *server) Discover(req *pb.DiscoverRequest, stream pb.DiscoveryHandler_DiscoverServer) error {
	// Implement your discovery logic here
	// For demonstration, let's just send a single device in response
	fmt.Println("gRPC call to server.Discover()")
	fmt.Println(req.DiscoveryDetails)
	discoveryDetails, err := NewDiscoveryDetails(req.DiscoveryDetails)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	fmt.Println("parsed discovery details")

	for {
		successIPs, err := discoveryDetails.Scan()
		if err != nil {
			fmt.Println("error scanning ip range: ", err.Error())
			return err
		}

		if len(successIPs) == 0 {
			fmt.Println("No devices discovered")
		} else {
			fmt.Println("got devices!")
		}

		var devices []*pb.Device
		for _, deviceIp := range successIPs {
			devices = append(devices, toProtobufDevice(deviceIp))
		}

		// if len(devices) == 0 {
		// 	fmt.Println("No devices discovered")
		// 	return nil
		// }
		fmt.Printf("Sending %d devices...\n", len(devices))

		res := &pb.DiscoverResponse{
			Devices: devices,
		}
		if err := stream.Send(res); err != nil {
			// if this errors out, we should re-register!
			fmt.Println("error sending streaming devices")
			fmt.Println(err.Error())

			fmt.Println("sending message to re-register")
			registerChan <- true
			return err
		}

		if len(devices) == 0 {
			time.Sleep(15 * time.Second)
			continue
		}

		time.Sleep(60 * time.Second)
	}
}

func toProtobufDevice(input string) *pb.Device {
	return &pb.Device{
		Id: input,
		Properties: map[string]string{
			"AKRI_HTTP":     "http",
			"HOST_ENDPOINT": input,
		},
		Mounts:      []*pb.Mount{},
		DeviceSpecs: []*pb.DeviceSpec{},
	}
}
