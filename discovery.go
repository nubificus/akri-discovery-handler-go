package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/smirzaei/parallel"
	"gopkg.in/yaml.v3"
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
	IPStart     string `yaml:"ipStart"`
	IPEnd       string `yaml:"ipEnd"`
	Application string `yaml:"applicationType"`
	Secure      bool   `yaml:"secure"`
}

func (d *DiscoveryDetails) UnmarshalYAML(value *yaml.Node) error {
	type rawDetails DiscoveryDetails
	raw := rawDetails{
		Secure: true, // Set default value if secure is not provided
	}
	if err := value.Decode(&raw); err != nil {
		return err
	}
	*d = DiscoveryDetails(raw)
	return nil
}

func NewDiscoveryDetails(input string) (DiscoveryDetails, error) {
	var details DiscoveryDetails
	err := yaml.Unmarshal([]byte(input), &details)
	if err != nil {
		return details, err
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
		target := fmt.Sprintf("%s%d", prefix, i)
		ips = append(ips, target)
	}
	return ips, nil
}

type DiscoveryRequest struct {
	IP         string
	DeviceType string
}

// DeviceState keeps track of probe status and retry timing.
type DeviceState struct {
	LastSeen     time.Time
	NextProbeDue time.Time
	Backoff      time.Duration
	Discovered   bool
}

var (
	deviceStates   = make(map[string]*DeviceState)
	deviceStatesMu sync.Mutex
)

func (details DiscoveryDetails) Scan() ([]Device, error) {
	maxConcurrency := 10
	ipAddresses, err := details.IPList()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	resultIPs := parallel.MapLimit(ipAddresses, maxConcurrency, func(ip string) Device {
		deviceStatesMu.Lock()
		state, exists := deviceStates[ip]
		if !exists {
			state = &DeviceState{
				Backoff:      10 * time.Second,
				NextProbeDue: now,
			}
			deviceStates[ip] = state
		}
		if now.Before(state.NextProbeDue) {
			deviceStatesMu.Unlock()
			return Device{Hostname: ip, Discovered: state.Discovered}
		}
		deviceStatesMu.Unlock()

		res := scanDevice(ip, details.Application, details.Secure)

		deviceStatesMu.Lock()
		defer deviceStatesMu.Unlock()
		if res.Discovered {
			state.LastSeen = now
			state.Backoff = 10 * time.Second
			state.NextProbeDue = now.Add(state.Backoff)
			state.Discovered = true
		} else {
			// exponential backoff capped at 5 min
			if state.Backoff < 5*time.Minute {
				state.Backoff *= 2
				if state.Backoff > 5*time.Minute {
					state.Backoff = 5 * time.Minute
				}
			}
			state.NextProbeDue = now.Add(state.Backoff)
			state.Discovered = false
		}
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

func scanDevice(ip string, expected string, secure bool) Device {
	dev := Device{
		Hostname:   ip,
		Discovered: false,
	}

	url := fmt.Sprintf("http://%s/info", ip)
	client := http.Client{
		Timeout: 3 * time.Second,
		Transport: &http.Transport{
			DisableKeepAlives: true,
		},
	}
	resp, err := client.Get(url)
	if err != nil {
		return dev
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return dev
	}
	var deviceInfo DeviceInfo
	err = json.Unmarshal(body, &deviceInfo)
	if err != nil {
		return dev
	}
	if deviceInfo.Application != expected {
		return dev
	}

	if secure {
		trusted := VerifyDevice(ip)
		if !trusted {
			fmt.Printf("Device %s is not trusted\n", ip)
			return dev
		}
		dev.Discovered = true
		dev.Info = deviceInfo
		return dev
	}

	dev.Discovered = true
	dev.Info = deviceInfo
	return dev
}
