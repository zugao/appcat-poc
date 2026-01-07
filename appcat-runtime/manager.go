package main

import (
	"context"
	"fmt"
	"time"

	fnv1 "github.com/crossplane/function-sdk-go/proto/v1"
	"github.com/go-logr/logr"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/durationpb"
)

// Manager handles composition function requests
type Manager struct {
	fnv1.UnimplementedFunctionRunnerServiceServer
	log           logr.Logger
	proxyEndpoint string
}

// NewManager creates a new Manager instance
func NewManager(log logr.Logger, proxyEndpoint string) *Manager {
	return &Manager{
		log:           log,
		proxyEndpoint: proxyEndpoint,
	}
}

// RunFunction implements the FunctionRunnerServiceServer interface
// Merges service config (defaultHelmValues + mapping) with user runtime parameters
func (m *Manager) RunFunction(ctx context.Context, req *fnv1.RunFunctionRequest) (*fnv1.RunFunctionResponse, error) {
	log := m.log.WithValues("function", "appcat-poc")

	// If proxy endpoint is set, forward request to local endpoint
	if m.proxyEndpoint != "" {
		log.Info("Proxy mode enabled - forwarding request", "endpoint", m.proxyEndpoint)
		return m.proxyFunction(ctx, req)
	}

	log.Info("RunFunction called")

	// STEP 1: Extract composite (contains user runtime parameters from XRD spec)
	composite := req.GetObserved().GetComposite()
	if composite == nil {
		return nil, fmt.Errorf("composite is nil")
	}

	userSpec, err := extractUserSpec(composite)
	if err != nil {
		log.Error(err, "Failed to extract user spec from composite")
		return nil, fmt.Errorf("failed to extract user spec: %w", err)
	}

	log.Info("Extracted user spec", "spec", userSpec)

	// STEP 2: Extract service config from Composition input
	input := req.GetInput()
	if input == nil {
		return nil, fmt.Errorf("input is nil")
	}

	serviceConfig, err := extractServiceConfig(input)
	if err != nil {
		log.Error(err, "Failed to extract service config from input")
		return nil, fmt.Errorf("failed to extract service config: %w", err)
	}

	log.Info("Extracted service config",
		"chartName", serviceConfig["chart"].(map[string]interface{})["name"],
		"hasMappings", len(serviceConfig["mapping"].(map[string]interface{})))

	// STEP 3: Merge configs
	// - Start with service defaultHelmValues
	// - Use mapping to inject user spec values into helm values
	mergedConfig, err := mergeConfigs(serviceConfig, userSpec, log)
	if err != nil {
		log.Error(err, "Failed to merge configs")
		return nil, fmt.Errorf("failed to merge configs: %w", err)
	}

	log.Info("Config merged successfully")

	// STEP 4: Generate desired resources
	resources, connDetails, err := generateResources(ctx, composite, req.GetObserved().GetResources(), mergedConfig, log)
	if err != nil {
		log.Error(err, "Failed to generate resources")
		return nil, fmt.Errorf("failed to generate resources: %w", err)
	}

	// STEP 5: Build and return response
	resp := &fnv1.RunFunctionResponse{
		Meta: &fnv1.ResponseMeta{
			Ttl: durationpb.New(60 * time.Second),
		},
		Desired: &fnv1.State{
			Composite: &fnv1.Resource{
				ConnectionDetails: connDetails,
				Ready:             fnv1.Ready_READY_TRUE,
			},
			Resources: resources,
		},
	}

	log.Info("Function execution complete", "resourceCount", len(resources))
	return resp, nil
}

// proxyFunction forwards the function request to a local development endpoint
// This allows debugging the function locally while the proxy runs in the cluster
func (m *Manager) proxyFunction(ctx context.Context, req *fnv1.RunFunctionRequest) (*fnv1.RunFunctionResponse, error) {
	log := m.log.WithValues("function", "appcat-poc-proxy")

	log.Info("Forwarding request to local endpoint", "endpoint", m.proxyEndpoint)

	// Create insecure gRPC connection to local endpoint
	// Local endpoint runs with proper TLS, but proxy connects without TLS for simplicity
	conn, err := grpc.DialContext(ctx, m.proxyEndpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		log.Error(err, "Failed to connect to proxy endpoint", "endpoint", m.proxyEndpoint)
		return nil, fmt.Errorf("failed to connect to proxy endpoint %s: %w", m.proxyEndpoint, err)
	}
	defer conn.Close()

	// Forward the request to the local function
	client := fnv1.NewFunctionRunnerServiceClient(conn)
	resp, err := client.RunFunction(ctx, req)
	if err != nil {
		log.Error(err, "Failed to execute function on proxy endpoint", "endpoint", m.proxyEndpoint)
		return nil, fmt.Errorf("failed to execute function on proxy endpoint %s: %w", m.proxyEndpoint, err)
	}

	log.Info("Successfully proxied request to local endpoint", "endpoint", m.proxyEndpoint)
	return resp, nil
}
