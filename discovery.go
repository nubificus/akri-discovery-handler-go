package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/smirzaei/parallel"
)

type Device struct {
	Hostname   string
	Discovered bool
}
type DiscoveredDevices struct {
	sync.Mutex
	Devices []string
}

type DiscoveryDetails struct {
	IPStart    string
	IPEnd      string
	DeviceType string
}

func NewDiscoveryDetails(input string) (DiscoveryDetails, error) {
	details := DiscoveryDetails{}
	lines := strings.Split(input, "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ": ", 2)
		if len(parts) != 2 {
			return details, errors.New("invalid input format")
		}
		switch parts[0] {
		case "ipStart":
			details.IPStart = parts[1]
		case "ipEnd":
			details.IPEnd = parts[1]
		case "deviceType":
			details.DeviceType = parts[1]
		default:
			return details, errors.New("unknown field in input")
		}
	}
	if details.IPStart == "" || details.IPEnd == "" || details.DeviceType == "" {
		return details, errors.New("missing fields in input")
	}
	return details, nil
}

func (details DiscoveryDetails) IPList() ([]string, error) {
	var ips []string
	parts := strings.Split(details.IPStart, ".")
	prefix := fmt.Sprintf("%s.%s.%s.", parts[0], parts[1], parts[2])
	start := parts[len(parts)-1]
	startInt, err := strconv.Atoi(start)
	if err != nil {
		return nil, err
	}
	parts = strings.Split(details.IPEnd, ".")
	end := parts[len(parts)-1]
	endInt, err := strconv.Atoi(end)
	if err != nil {
		return nil, err
	}
	for i := startInt; i <= endInt; i++ {
		// target := fmt.Sprintf("http://%s%d/device", prefix, i)
		target := fmt.Sprintf("%s%d", prefix, i)
		ips = append(ips, target)
	}
	return ips, nil
}

type DiscoveryRequest struct {
	IP         string
	DeviceType string
}

func (details DiscoveryDetails) Scan() ([]string, error) {
	maxConcurrency := 10
	ipAddresses, err := details.IPList()
	if err != nil {
		return nil, err
	}

	resultIPs := parallel.MapLimit(ipAddresses, maxConcurrency, func(ip string) Device {
		res := scanDevice(ip, details.DeviceType)
		return res
	})
	successIPs := []string{}
	for _, res := range resultIPs {
		if res.Discovered {
			successIPs = append(successIPs, res.Hostname)
		}
	}
	fmt.Println("Scanned devices:", successIPs)
	return successIPs, nil
}

func scanDevice(ip string, expected string) Device {
	dev := Device{
		Hostname:   ip,
		Discovered: false,
	}
	url := fmt.Sprintf("http://%s/device", ip)
	client := http.Client{
		Timeout: 3 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		// fmt.Printf("Error fetching %s: %v\n", url, err)
		return dev
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		// fmt.Printf("Error reading response body from %s: %v\n", url, err)
		return dev
	}
	if string(body) == expected {
		dev.Discovered = true
	}
	return dev
}
