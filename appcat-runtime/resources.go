package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

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

// getSecretName extracts secret name from writeConnectionSecretToRef or falls back to composite name
func getSecretName(composite *fnv1.Resource, compositeNamespace string, log logr.Logger) (string, string, error) {
	compositeMap := composite.Resource.AsMap()
	paved := fieldpath.Pave(compositeMap)

	// Get composite name as fallback
	compositeName, err := paved.GetString("metadata.name")
	if err != nil {
		return "", "", fmt.Errorf("failed to get composite name: %w", err)
	}

	// Try to extract writeConnectionSecretToRef
	writeSecretRef, err := paved.GetValue("spec.writeConnectionSecretToRef")
	if err != nil || writeSecretRef == nil {
		log.Info("No writeConnectionSecretToRef, using composite name", "name", compositeName)
		return compositeName, compositeNamespace, nil
	}

	secretRef, ok := writeSecretRef.(map[string]interface{})
	if !ok {
		return compositeName, compositeNamespace, nil
	}

	// Extract name and namespace
	secretName := compositeName       // fallback
	secretNamespace := compositeNamespace // fallback

	if name, ok := secretRef["name"].(string); ok && name != "" {
		secretName = name
	}

	if ns, ok := secretRef["namespace"].(string); ok && ns != "" {
		secretNamespace = ns
	}

	log.Info("Using secret configuration",
		"secretName", secretName,
		"secretNamespace", secretNamespace)

	return secretName, secretNamespace, nil
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

	// Extract connection secret configuration from input
	connectionSecret, err := getConnectionSecretConfig(mergedConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get connection secret config: %w", err)
	}

	// Get secret name early so we can inject it into Helm values if needed
	secretName, secretNamespace, err := getSecretName(composite, compositeNamespace, log)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get secret name: %w", err)
	}

	// Inject password into Helm values if passwordPath specified
	if connectionSecret.PasswordPath != "" {
		if err := injectPasswordIntoHelmValues(helmValues, connectionSecret.PasswordPath, password); err != nil {
			return nil, nil, fmt.Errorf("failed to inject password into Helm values: %w", err)
		}
	}

	// Inject existing secret name into Helm values if existingSecretPath specified
	if connectionSecret.ExistingSecretPath != "" {
		paved := fieldpath.Pave(helmValues)
		if err := paved.SetValue(connectionSecret.ExistingSecretPath, secretName); err != nil {
			return nil, nil, fmt.Errorf("failed to inject secret name into Helm values: %w", err)
		}
		log.Info("Configured Helm chart to use existing secret",
			"path", connectionSecret.ExistingSecretPath,
			"secretName", secretName)
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

	// Build variable map for template substitution
	variables := map[string]string{
		"instanceName": instanceName,
		"namespace":    compositeNamespace,
		"password":     password,
	}

	// Create Secret as managed resource

	// Generate connection details from templates
	connDetails := make(map[string][]byte)
	secretBuilder := NewSecretBuilder(secretName, secretNamespace)

	for _, field := range connectionSecret.Fields {
		// Substitute variables in template
		value := substituteVariables(field.Value, variables)
		connDetails[field.Key] = []byte(value)
		secretBuilder = secretBuilder.WithData(field.Key, []byte(value))
	}

	log.Info("Creating connection Secret",
		"secretName", secretName,
		"secretNamespace", secretNamespace,
		"instanceName", instanceName,
		"fieldsCount", len(connectionSecret.Fields))

	secret := secretBuilder.
		WithLabel("app.kubernetes.io/managed-by", "crossplane").
		WithLabel("app.kubernetes.io/instance", instanceName).
		WithLabel("app.kubernetes.io/component", "connection-secret").
		Build()

	secretResource, err := toFunctionResource(secret)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to convert secret to function resource: %w", err)
	}
	resources["secret"] = secretResource

	log.Info("Generated all resources", "count", len(resources))
	return resources, connDetails, nil
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

// SecretFieldTemplate represents a single secret field with templated value
type SecretFieldTemplate struct {
	Key   string
	Value string
}

// ConnectionSecretConfig defines the structure and content of connection secrets
type ConnectionSecretConfig struct {
	Fields             []SecretFieldTemplate
	PasswordPath       string
	ExistingSecretPath string
}

// getConnectionSecretConfig extracts connectionSecret configuration from merged config
func getConnectionSecretConfig(mergedConfig map[string]interface{}) (*ConnectionSecretConfig, error) {
	secretConfig, ok := mergedConfig["connectionSecret"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("connectionSecret not found in merged config")
	}

	// Parse fields array
	fields := []SecretFieldTemplate{}
	if fieldsRaw, ok := secretConfig["fields"].([]interface{}); ok {
		for _, fieldRaw := range fieldsRaw {
			if fieldMap, ok := fieldRaw.(map[string]interface{}); ok {
				key, _ := fieldMap["key"].(string)
				value, _ := fieldMap["value"].(string)
				fields = append(fields, SecretFieldTemplate{Key: key, Value: value})
			}
		}
	}

	passwordPath, _ := secretConfig["passwordPath"].(string)
	existingSecretPath, _ := secretConfig["existingSecretPath"].(string)

	return &ConnectionSecretConfig{
		Fields:             fields,
		PasswordPath:       passwordPath,
		ExistingSecretPath: existingSecretPath,
	}, nil
}

// substituteVariables performs ${var} substitution in template strings
// Supported variables: ${instanceName}, ${namespace}, ${password}
func substituteVariables(template string, variables map[string]string) string {
	result := template
	for key, value := range variables {
		placeholder := fmt.Sprintf("${%s}", key)
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}

// injectPasswordIntoHelmValues injects password at specified path using dot notation
func injectPasswordIntoHelmValues(helmValues map[string]interface{}, path string, password string) error {
	if path == "" {
		return nil // No password path specified, skip injection
	}
	paved := fieldpath.Pave(helmValues)
	return paved.SetValue(path, password)
}
