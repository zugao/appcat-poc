package main

import (
	"context"
	"fmt"

	fnv1 "github.com/crossplane/function-sdk-go/proto/v1"
	"github.com/crossplane/crossplane-runtime/pkg/fieldpath"
	"github.com/go-logr/logr"
	"google.golang.org/protobuf/types/known/structpb"
	"k8s.io/apimachinery/pkg/runtime"
)

// generateResources creates the desired Kubernetes resources
// Returns: Namespace, Secret, HelmRelease
func generateResources(
	ctx context.Context,
	composite *fnv1.Resource,
	mergedConfig map[string]interface{},
	log logr.Logger,
) (map[string]*fnv1.Resource, error) {
	// Extract instance name from composite metadata
	compositeMap := composite.Resource.AsMap()
	paved := fieldpath.Pave(compositeMap)

	instanceName, err := paved.GetString("metadata.name")
	if err != nil {
		return nil, fmt.Errorf("failed to get instance name: %w", err)
	}

	namespace := fmt.Sprintf("vshn-redis-%s", instanceName)
	log.Info("Generating resources", "instance", instanceName, "namespace", namespace)

	resources := make(map[string]*fnv1.Resource)

	// 1. Generate Namespace
	ns := NewNamespaceBuilder(namespace).
		WithLabel("app", "redis").
		WithLabel("instance", instanceName).
		Build()

	nsResource, err := toFunctionResource(ns)
	if err != nil {
		return nil, fmt.Errorf("failed to convert namespace to function resource: %w", err)
	}
	resources["namespace"] = nsResource

	// 2. Generate Secret (for Redis password)
	secret := NewSecretBuilder(fmt.Sprintf("%s-password", instanceName), namespace).
		WithRandomPassword("password", 32).
		Build()

	secretResource, err := toFunctionResource(secret)
	if err != nil {
		return nil, fmt.Errorf("failed to convert secret to function resource: %w", err)
	}
	resources["secret"] = secretResource

	// 3. Generate HelmRelease
	chart, ok := mergedConfig["chart"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("chart not found in merged config")
	}

	chartName, ok := chart["name"].(string)
	if !ok {
		return nil, fmt.Errorf("chart.name not found or not a string")
	}

	chartRepo, ok := chart["repository"].(string)
	if !ok {
		return nil, fmt.Errorf("chart.repository not found or not a string")
	}

	chartVersion, ok := chart["defaultVersion"].(string)
	if !ok {
		return nil, fmt.Errorf("chart.defaultVersion not found or not a string")
	}

	helmValues, ok := mergedConfig["helmValues"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("helmValues not found in merged config")
	}

	// Inject secret reference into helm values
	helmValues["auth"] = map[string]interface{}{
		"enabled":                    true,
		"existingSecret":             fmt.Sprintf("%s-password", instanceName),
		"existingSecretPasswordKey":  "password",
	}

	log.Info("Creating HelmRelease",
		"chart", chartName,
		"version", chartVersion,
		"repository", chartRepo)

	helmRelease := NewHelmReleaseBuilder(instanceName).
		WithChart(chartRepo, chartName, chartVersion).
		WithTargetNamespace(namespace).
		WithValues(helmValues).
		Build()

	helmReleaseResource, err := toFunctionResource(helmRelease)
	if err != nil {
		return nil, fmt.Errorf("failed to convert helm release to function resource: %w", err)
	}
	resources["helmrelease"] = helmReleaseResource

	log.Info("Generated all resources", "count", len(resources))
	return resources, nil
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