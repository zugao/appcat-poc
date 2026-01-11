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
		return nil, fmt.Errorf("failed to extract service config: %w", err)
	}
	log.Info("Extracted service config")

	// STEP 3: Merge configs (defaultHelmValues + user parameters)
	mergedConfig, err := mergeConfigs(serviceConfig, userSpec, log)
	if err != nil {
		return nil, fmt.Errorf("failed to merge configs: %w", err)
	}

	// STEP 4: Generate desired resources
	resources, connDetails, err := generateResources(ctx, composite, req.GetObserved().GetResources(), mergedConfig, log)
	if err != nil {
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

// proxyFunction forwards requests to a local development endpoint for debugging
func (m *Manager) proxyFunction(ctx context.Context, req *fnv1.RunFunctionRequest) (*fnv1.RunFunctionResponse, error) {
	// Create insecure gRPC connection to local endpoint
	conn, err := grpc.NewClient(m.proxyEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to proxy %s: %w", m.proxyEndpoint, err)
	}
	defer conn.Close()

	// Forward the request
	client := fnv1.NewFunctionRunnerServiceClient(conn)
	resp, err := client.RunFunction(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute on proxy %s: %w", m.proxyEndpoint, err)
	}

	return resp, nil
}
