package main

import (
	"context"
	"fmt"
	"time"

	fnv1 "github.com/crossplane/function-sdk-go/proto/v1"
	"github.com/go-logr/logr"
	"google.golang.org/protobuf/types/known/durationpb"
)

// Manager handles composition function requests
type Manager struct {
	fnv1.UnimplementedFunctionRunnerServiceServer
	log logr.Logger
}

// NewManager creates a new Manager instance
func NewManager(log logr.Logger) *Manager {
	return &Manager{
		log: log,
	}
}

// RunFunction implements the FunctionRunnerServiceServer interface
// Merges service config (defaultHelmValues + mapping) with user runtime parameters
func (m *Manager) RunFunction(ctx context.Context, req *fnv1.RunFunctionRequest) (*fnv1.RunFunctionResponse, error) {
	log := m.log.WithValues("function", "appcat-poc")
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
	resources, _, err := generateResources(ctx, composite, req.GetObserved().GetResources(), mergedConfig, log)
	if err != nil {
		log.Error(err, "Failed to generate resources")
		return nil, fmt.Errorf("failed to generate resources: %w", err)
	}

	// STEP 5: Build and return response
	// Note: In Crossplane 2.x, connection details are managed by creating a Secret
	// as a composed resource (not via Composite.ConnectionDetails)
	resp := &fnv1.RunFunctionResponse{
		Meta: &fnv1.ResponseMeta{
			Ttl: durationpb.New(60 * time.Second),
		},
		Desired: &fnv1.State{
			Resources: resources, // Includes: helmrelease + connection-secret
		},
	}

	log.Info("Function execution complete", "resourceCount", len(resources))
	return resp, nil
}