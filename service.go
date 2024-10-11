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
		fmt.Println("entered discovery loop")
		startTime := time.Now()
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
		for _, device := range successIPs {
			devices = append(devices, toProtobufDevice(device))
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
		endTime := time.Now()
		elapsedTime := endTime.Sub(startTime).Seconds()
		fmt.Printf("Elapsed time: %.2f seconds\n", elapsedTime)

		if len(devices) == 0 {
			fmt.Println("No devices were discovered. Sleeping for 15 seconds")

			time.Sleep(15 * time.Second)
			continue
		}
		fmt.Println("devices returned successfully! sleeping for 30 seconds")
		time.Sleep(30 * time.Second)
	}
}

func toProtobufDevice(input Device) *pb.Device {
	return &pb.Device{
		Id: input.Hostname,
		Properties: map[string]string{
			"AKRI_HTTP":        "http",
			"HOST_ENDPOINT":    input.Hostname,
			"APPLICATION_TYPE": input.Info.Application,
			"DEVICE":           input.Info.Device,
			"VERSION":          input.Info.Version,
		},
		Mounts:      []*pb.Mount{},
		DeviceSpecs: []*pb.DeviceSpec{},
	}
}
