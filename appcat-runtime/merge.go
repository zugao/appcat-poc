package main

import (
	"fmt"
	"strings"

	fnv1 "github.com/crossplane/function-sdk-go/proto/v1"
	"github.com/crossplane/crossplane-runtime/pkg/fieldpath"
	"github.com/go-logr/logr"
	"google.golang.org/protobuf/types/known/structpb"
)

// extractUserSpec extracts user-provided spec from the composite resource
// Returns a map with the full spec (e.g., {size: {cpu: "1000m"}, replicas: 3})
func extractUserSpec(composite *fnv1.Resource) (map[string]interface{}, error) {
	compositeMap := composite.Resource.AsMap()
	paved := fieldpath.Pave(compositeMap)

	specRaw, err := paved.GetValue("spec")
	if err != nil {
		return nil, fmt.Errorf("failed to get spec from composite: %w", err)
	}

	spec, ok := specRaw.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("spec is not a map[string]interface{}")
	}

	return spec, nil
}

// extractServiceConfig extracts service configuration from Composition input
// Returns a map with: chart, defaultHelmValues, mapping, connectionSecret
func extractServiceConfig(input *structpb.Struct) (map[string]interface{}, error) {
	inputMap := input.AsMap()
	paved := fieldpath.Pave(inputMap)

	dataRaw, err := paved.GetValue("data")
	if err != nil {
		return nil, fmt.Errorf("failed to get data from input: %w", err)
	}

	data, ok := dataRaw.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("data is not a map[string]interface{}")
	}

	// Validate required fields
	if _, ok := data["chart"]; !ok {
		return nil, fmt.Errorf("chart not found in service config")
	}
	if _, ok := data["defaultHelmValues"]; !ok {
		return nil, fmt.Errorf("defaultHelmValues not found in service config")
	}
	if _, ok := data["mapping"]; !ok {
		return nil, fmt.Errorf("mapping not found in service config")
	}
	if _, ok := data["connectionSecret"]; !ok {
		return nil, fmt.Errorf("connectionSecret not found in service config")
	}

	return data, nil
}

// mergeConfigs merges service config with user spec using the provided mapping
// Returns a merged config with: chart, helmValues (merged), connectionSecret
func mergeConfigs(serviceConfig map[string]interface{}, userSpec map[string]interface{}, log logr.Logger) (map[string]interface{}, error) {
	// Start with service's defaultHelmValues (deep copy)
	defaultHelmValues, ok := serviceConfig["defaultHelmValues"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("defaultHelmValues is not a map")
	}
	helmValues := deepCopy(defaultHelmValues)

	// Get mapping
	mapping, ok := serviceConfig["mapping"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("mapping is not a map")
	}

	// Apply mappings: inject user spec values into helm values
	for xrdPath, helmPathRaw := range mapping {
		helmPath, ok := helmPathRaw.(string)
		if !ok {
			log.Info("Skipping non-string helm path", "xrdPath", xrdPath, "helmPath", helmPathRaw)
			continue
		}

		// Get value from user spec using XRD path
		value, err := getValueByPath(userSpec, xrdPath)
		if err != nil {
			// User didn't provide this field - skip it
			log.Info("User spec doesn't have value for path", "xrdPath", xrdPath)
			continue
		}

		// Set value in helm values using helm path
		err = setValueByPath(helmValues, helmPath, value)
		if err != nil {
			log.Error(err, "Failed to set helm value", "helmPath", helmPath, "value", value)
			return nil, fmt.Errorf("failed to set helm value at path %s: %w", helmPath, err)
		}

		log.Info("Mapped value", "xrdPath", xrdPath, "helmPath", helmPath, "value", value)
	}

	// Return merged config
	result := map[string]interface{}{
		"chart":      serviceConfig["chart"],
		"helmValues": helmValues,
	}

	// Include connectionSecret if present in service config
	if connectionSecret, ok := serviceConfig["connectionSecret"]; ok {
		result["connectionSecret"] = connectionSecret
	}

	return result, nil
}

// getValueByPath retrieves a value from a nested map using a dot-separated path
// Example: "spec.size.cpu" -> userSpec["size"]["cpu"]
func getValueByPath(data map[string]interface{}, path string) (interface{}, error) {
	parts := strings.Split(path, ".")
	current := interface{}(data)

	for i, part := range parts {
		// Skip "spec" prefix if present
		if i == 0 && part == "spec" {
			continue
		}

		m, ok := current.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("path %s: expected map at part %s, got %T", path, part, current)
		}

		value, ok := m[part]
		if !ok {
			return nil, fmt.Errorf("path %s: key %s not found", path, part)
		}

		current = value
	}

	return current, nil
}

// setValueByPath sets a value in a nested map using a dot-separated path
// Creates intermediate maps if they don't exist
// Example: "master.resources.requests.cpu" with value "1000m"
func setValueByPath(data map[string]interface{}, path string, value interface{}) error {
	parts := strings.Split(path, ".")
	if len(parts) == 0 {
		return fmt.Errorf("empty path")
	}

	// Navigate to the parent of the final key, creating maps as needed
	current := data
	for i := 0; i < len(parts)-1; i++ {
		part := parts[i]

		// Get or create the next level
		next, ok := current[part]
		if !ok {
			// Create new map
			next = make(map[string]interface{})
			current[part] = next
		}

		// Ensure it's a map
		nextMap, ok := next.(map[string]interface{})
		if !ok {
			return fmt.Errorf("path %s: expected map at part %s, got %T", path, part, next)
		}

		current = nextMap
	}

	// Set the final value
	finalKey := parts[len(parts)-1]
	current[finalKey] = value

	return nil
}

// deepCopy creates a deep copy of a map[string]interface{}
func deepCopy(src map[string]interface{}) map[string]interface{} {
	dst := make(map[string]interface{})
	for k, v := range src {
		switch val := v.(type) {
		case map[string]interface{}:
			dst[k] = deepCopy(val)
		case []interface{}:
			dst[k] = deepCopySlice(val)
		default:
			dst[k] = v
		}
	}
	return dst
}

// deepCopySlice creates a deep copy of a []interface{}
func deepCopySlice(src []interface{}) []interface{} {
	dst := make([]interface{}, len(src))
	for i, v := range src {
		switch val := v.(type) {
		case map[string]interface{}:
			dst[i] = deepCopy(val)
		case []interface{}:
			dst[i] = deepCopySlice(val)
		default:
			dst[i] = v
		}
	}
	return dst
}