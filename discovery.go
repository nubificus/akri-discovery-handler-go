package main

import (
	"encoding/json"
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
	Info       DeviceInfo
}
type DiscoveredDevices struct {
	sync.Mutex
	Devices []string
}

type DiscoveryDetails struct {
	IPStart     string
	IPEnd       string
	Application string
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
		case "applicationType":
			details.Application = parts[1]
		default:
			return details, errors.New("unknown field in input")
		}
	}
	if details.IPStart == "" || details.IPEnd == "" || details.Application == "" {
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

func (details DiscoveryDetails) Scan() ([]Device, error) {
	maxConcurrency := 10
	ipAddresses, err := details.IPList()
	if err != nil {
		return nil, err
	}

	resultIPs := parallel.MapLimit(ipAddresses, maxConcurrency, func(ip string) Device {
		res := scanDevice(ip, details.Application)
		return res
	})
	successIPs := []Device{}
	for _, res := range resultIPs {
		if res.Discovered {
			successIPs = append(successIPs, res)
		}
	}
	fmt.Println("Scanned devices:", successIPs)
	return successIPs, nil
}

type DeviceInfo struct {
	Device      string `json:"device"`
	Application string `json:"application"`
	Version     string `json:"version"`
}

func scanDevice(ip string, expected string) Device {
	dev := Device{
		Hostname:   ip,
		Discovered: false,
	}

	url := fmt.Sprintf("http://%s/info", ip)
	client := http.Client{
		Timeout: 3 * time.Second,
		Transport: &http.Transport{
			DisableKeepAlives: true, // Prevents connection reuse
		},
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
	// Parse the JSON response
	var deviceInfo DeviceInfo
	err = json.Unmarshal(body, &deviceInfo)
	if err != nil {
		// fmt.Printf("Error parsing JSON from %s: %v\n", url, err)
		return dev
	}
	if deviceInfo.Application == expected {
		fmt.Printf("Discovered device %s with type %s\n", ip, deviceInfo.Application)
		dev.Discovered = true
		dev.Info = deviceInfo // Store the parsed information
	}
	return dev
}
