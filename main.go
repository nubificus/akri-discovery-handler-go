package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/nubificus/akri-discovery-handler-go/pkg/pb"
	"github.com/nubificus/akri-discovery-handler-go/pkg/utils"
	"google.golang.org/grpc"
)

const agentRegistrationSocketName = "agent-registration.sock"
const discoveryHandlerPrefix = "go-http-range"

var (
	agentRegistrationSocketPath string
	discoveryHandlerName        string
	discoveryServiceSocketPath  string
)

func init() {
	envVar := os.Getenv("DISCOVERY_HANDLER_SUFFIX")
	discoveryHandlerName = fmt.Sprintf("%s%s", discoveryHandlerPrefix, envVar)

	envVar = os.Getenv("DISCOVERY_HANDLERS_DIRECTORY")
	if envVar == "" {
		envVar = "/var/lib/akri"
	}
	agentRegistrationSocketPath = filepath.Join(envVar, agentRegistrationSocketName)
	discoveryServiceSocketPath = filepath.Join(envVar, fmt.Sprintf("%s.sock", discoveryHandlerName))
	os.RemoveAll(discoveryServiceSocketPath)
}

func main() {
	fmt.Println("go-http-range started")
	fmt.Println("Agent Registration Socket: ", agentRegistrationSocketPath)
	fmt.Println("Discovery Handler Complete Name: ", discoveryHandlerName)
	fmt.Println("Discovery Service Socket: ", discoveryServiceSocketPath)
	dumpEnvs()

	var wg sync.WaitGroup
	// Register with agent
	wg.Add(1)
	go func() {
		defer wg.Done()
		// periodically re-register
		for {
			err := utils.RegisterDiscoveryHandler(agentRegistrationSocketPath, discoveryHandlerName, discoveryServiceSocketPath)
			if err != nil {
				log.Fatal(err.Error())
			}
			time.Sleep(20 * time.Second)
		}

	}()
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	wg.Add(1)
	go func() {
		defer wg.Done()
		// Start Service
		lis, err := net.Listen("unix", discoveryServiceSocketPath)
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		defer os.Remove(discoveryServiceSocketPath)
		newServer := &server{}
		pb.RegisterDiscoveryHandlerServer(grpcServer, newServer)
		log.Println("Server started at", discoveryServiceSocketPath)
		fmt.Println("Started service")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()
	// Handle OS signals for graceful shutdown
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	log.Println("Shutting down server...")
	grpcServer.GracefulStop()
	log.Println("Server gracefully stopped")
	// wg.Wait()
}

func dumpEnvs() {
	fmt.Println("Environment variables:")
	env := os.Environ()
	envMap := make(map[string]string)
	for _, v := range env {
		pair := strings.SplitN(v, "=", 2)
		if len(pair) == 2 {
			envMap[pair[0]] = pair[1]
		}
	}
	for key, value := range envMap {
		fmt.Printf("%s=%s\n", key, value)
	}
}
