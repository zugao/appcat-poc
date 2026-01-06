package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/crossplane/crossplane-runtime/pkg/fieldpath"
	fnv1 "github.com/crossplane/function-sdk-go/proto/v1"
	"github.com/go-logr/logr"
	"google.golang.org/protobuf/types/known/structpb"
	"k8s.io/apimachinery/pkg/runtime"
)

// getOrGeneratePassword retrieves existing password from observed HelmRelease or generates new one
func getOrGeneratePassword(observedResources map[string]*fnv1.Resource, instanceName string, log logr.Logger) (string, error) {
	// Check for existing HelmRelease in observed resources
	if helmResource, exists := observedResources["helmrelease"]; exists && helmResource != nil {
		helmMap := helmResource.Resource.AsMap()
		paved := fieldpath.Pave(helmMap)

		// Try to extract existing password from HelmRelease values
		if passwordRaw, err := paved.GetValue("spec.forProvider.values.auth.password"); err == nil {
			if password, ok := passwordRaw.(string); ok && password != "" {
				log.Info("Reusing existing password from HelmRelease", "instance", instanceName)
				return password, nil
			}
		}
	}

	// No existing password - generate new one
	log.Info("Generating new password", "instance", instanceName)
	return generateRandomPassword(32), nil
}

// generateRandomPassword generates a random base64 password
func generateRandomPassword(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)[:length]
}

// generateResources creates the desired Kubernetes resources
// Returns: resources, connectionDetails, error
func generateResources(
	ctx context.Context,
	composite *fnv1.Resource,
	observedResources map[string]*fnv1.Resource,
	mergedConfig map[string]interface{},
	log logr.Logger,
) (map[string]*fnv1.Resource, map[string][]byte, error) {
	// Extract instance name from composite metadata
	compositeMap := composite.Resource.AsMap()
	paved := fieldpath.Pave(compositeMap)

	instanceName, err := paved.GetString("metadata.name")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get instance name: %w", err)
	}

	// Get composite namespace - all resources go in the same namespace for namespace-scoped composites
	compositeNamespace, err := paved.GetString("metadata.namespace")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get composite namespace: %w", err)
	}

	// For namespace-scoped Releases, the Helm chart deploys to the same namespace as the Release resource
	log.Info("Generating resources",
		"instance", instanceName,
		"compositeNamespace", compositeNamespace)

	resources := make(map[string]*fnv1.Resource)

	// 1. Generate Namespace - create as HelmRelease value, not as a managed resource
	// (Namespaces are cluster-scoped and cannot be composed by namespace-scoped composites)

	// 2. Get or generate password (will be published as connection detail)
	password, err := getOrGeneratePassword(observedResources, instanceName, log)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get password: %w", err)
	}

	// 3. Generate HelmRelease
	chart, ok := mergedConfig["chart"].(map[string]interface{})
	if !ok {
		return nil, nil, fmt.Errorf("chart not found in merged config")
	}

	chartName, ok := chart["name"].(string)
	if !ok {
		return nil, nil, fmt.Errorf("chart.name not found or not a string")
	}

	chartRepo, ok := chart["repository"].(string)
	if !ok {
		return nil, nil, fmt.Errorf("chart.repository not found or not a string")
	}

	chartVersion, ok := chart["defaultVersion"].(string)
	if !ok {
		return nil, nil, fmt.Errorf("chart.defaultVersion not found or not a string")
	}

	helmValues, ok := mergedConfig["helmValues"].(map[string]interface{})
	if !ok {
		return nil, nil, fmt.Errorf("helmValues not found in merged config")
	}

	// Inject password directly into Helm values
	helmValues["auth"] = map[string]interface{}{
		"enabled":  true,
		"password": password,
	}

	log.Info("Creating HelmRelease",
		"chart", chartName,
		"version", chartVersion,
		"repository", chartRepo,
		"helmReleaseNamespace", compositeNamespace)

	helmRelease := NewHelmReleaseBuilder(instanceName).
		WithNamespace(compositeNamespace).
		WithChart(chartRepo, chartName, chartVersion).
		WithValues(helmValues).
		Build()

	helmReleaseResource, err := toFunctionResource(helmRelease)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to convert helm release to function resource: %w", err)
	}
	resources["helmrelease"] = helmReleaseResource

	// Build connection details Secret
	connectionSecret := NewSecretBuilder(fmt.Sprintf("%s-connection", instanceName), compositeNamespace).
		WithStringData("password", password).
		WithStringData("host", fmt.Sprintf("%s-master.%s.svc.cluster.local", instanceName, compositeNamespace)).
		WithStringData("port", "6379").
		WithStringData("url", fmt.Sprintf("redis://:%s@%s-master.%s.svc.cluster.local:6379", password, instanceName, compositeNamespace)).
		WithLabel("app", "redis").
		WithLabel("instance", instanceName).
		Build()

	connectionSecretResource, err := toFunctionResource(connectionSecret)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to convert connection secret to function resource: %w", err)
	}
	resources["connection-secret"] = connectionSecretResource

	log.Info("Generated all resources", "count", len(resources))
	// Return empty connection details map since we're creating the secret as a composed resource
	return resources, map[string][]byte{}, nil
}

// toFunctionResource converts a Kubernetes runtime.Object to a function Resource
func toFunctionResource(obj runtime.Object) (*fnv1.Resource, error) {
	// Convert to unstructured
	unstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to unstructured: %w", err)
	}

	// Convert to structpb.Struct
	structpbStruct, err := structpb.NewStruct(unstructuredMap)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to structpb: %w", err)
	}

	return &fnv1.Resource{
		Resource: structpbStruct,
	}, nil
}
