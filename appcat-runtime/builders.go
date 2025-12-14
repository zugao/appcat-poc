package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"

	helmv1 "github.com/crossplane-contrib/provider-helm/apis/release/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// NamespaceBuilder builds Kubernetes Namespace objects using fluent API
type NamespaceBuilder struct {
	name        string
	labels      map[string]string
	annotations map[string]string
}

// NewNamespaceBuilder creates a new namespace builder
func NewNamespaceBuilder(name string) *NamespaceBuilder {
	return &NamespaceBuilder{
		name:        name,
		labels:      make(map[string]string),
		annotations: make(map[string]string),
	}
}

// WithLabel adds a label to the namespace
func (b *NamespaceBuilder) WithLabel(key, value string) *NamespaceBuilder {
	b.labels[key] = value
	return b
}

// WithLabels adds multiple labels to the namespace
func (b *NamespaceBuilder) WithLabels(labels map[string]string) *NamespaceBuilder {
	for k, v := range labels {
		b.labels[k] = v
	}
	return b
}

// WithAnnotation adds an annotation to the namespace
func (b *NamespaceBuilder) WithAnnotation(key, value string) *NamespaceBuilder {
	b.annotations[key] = value
	return b
}

// Build creates the Namespace object
func (b *NamespaceBuilder) Build() *corev1.Namespace {
	return &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Namespace",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        b.name,
			Labels:      b.labels,
			Annotations: b.annotations,
		},
	}
}

// SecretBuilder builds Kubernetes Secret objects using fluent API
type SecretBuilder struct {
	name        string
	namespace   string
	data        map[string][]byte
	stringData  map[string]string
	labels      map[string]string
	annotations map[string]string
}

// NewSecretBuilder creates a new secret builder
func NewSecretBuilder(name, namespace string) *SecretBuilder {
	return &SecretBuilder{
		name:        name,
		namespace:   namespace,
		data:        make(map[string][]byte),
		stringData:  make(map[string]string),
		labels:      make(map[string]string),
		annotations: make(map[string]string),
	}
}

// WithData adds binary data to the secret
func (b *SecretBuilder) WithData(key string, value []byte) *SecretBuilder {
	b.data[key] = value
	return b
}

// WithStringData adds string data to the secret
func (b *SecretBuilder) WithStringData(key, value string) *SecretBuilder {
	b.stringData[key] = value
	return b
}

// WithLabel adds a label to the secret
func (b *SecretBuilder) WithLabel(key, value string) *SecretBuilder {
	b.labels[key] = value
	return b
}

// WithLabels adds multiple labels to the secret
func (b *SecretBuilder) WithLabels(labels map[string]string) *SecretBuilder {
	for k, v := range labels {
		b.labels[k] = v
	}
	return b
}

// WithAnnotation adds an annotation to the secret
func (b *SecretBuilder) WithAnnotation(key, value string) *SecretBuilder {
	b.annotations[key] = value
	return b
}

// WithRandomPassword generates a random password and adds it to the secret
func (b *SecretBuilder) WithRandomPassword(key string, length int) *SecretBuilder {
	password := generateRandomPassword(length)
	b.stringData[key] = password
	return b
}

// Build creates the Secret object
func (b *SecretBuilder) Build() *corev1.Secret {
	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        b.name,
			Namespace:   b.namespace,
			Labels:      b.labels,
			Annotations: b.annotations,
		},
		Data:       b.data,
		StringData: b.stringData,
	}
}

// generateRandomPassword generates a random base64 password
func generateRandomPassword(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)[:length]
}

// ValuesFromConfig merges JSON helmValues from ConfigMap data with defaults.
func ValuesFromConfig(cfg map[string]string, defaults map[string]any) map[string]any {
	out := map[string]any{}
	for k, v := range defaults {
		out[k] = v
	}
	if raw, ok := cfg["helmValues"]; ok && raw != "" {
		_ = json.Unmarshal([]byte(raw), &out)
	}
	return out
}

// HelmReleaseBuilder builds helm.crossplane.io/v1beta1 Release objects using fluent API
// Note: HelmRelease is cluster-scoped, so it has no namespace in metadata.
// Use WithTargetNamespace() to specify where the chart deploys.
type HelmReleaseBuilder struct {
	name            string
	chartRepo       string
	chartName       string
	chartVersion    string
	targetNamespace string
	values          map[string]any
	labels          map[string]string
	annotations     map[string]string
}

// NewHelmReleaseBuilder creates a new HelmRelease builder
// Note: HelmRelease is cluster-scoped and has no namespace in metadata
func NewHelmReleaseBuilder(name string) *HelmReleaseBuilder {
	return &HelmReleaseBuilder{
		name:        name,
		values:      make(map[string]any),
		labels:      make(map[string]string),
		annotations: make(map[string]string),
	}
}

// WithChart sets the Helm chart repository, name, and version
func (b *HelmReleaseBuilder) WithChart(repo, name, version string) *HelmReleaseBuilder {
	b.chartRepo = repo
	b.chartName = name
	b.chartVersion = version
	return b
}

// WithTargetNamespace sets the namespace where the chart will be deployed
func (b *HelmReleaseBuilder) WithTargetNamespace(namespace string) *HelmReleaseBuilder {
	b.targetNamespace = namespace
	return b
}

// WithValues sets the Helm values
func (b *HelmReleaseBuilder) WithValues(values map[string]any) *HelmReleaseBuilder {
	b.values = values
	return b
}

// WithValue sets a single Helm value
func (b *HelmReleaseBuilder) WithValue(key string, value any) *HelmReleaseBuilder {
	b.values[key] = value
	return b
}

// WithLabel adds a label to the HelmRelease
func (b *HelmReleaseBuilder) WithLabel(key, value string) *HelmReleaseBuilder {
	b.labels[key] = value
	return b
}

// WithLabels adds multiple labels to the HelmRelease
func (b *HelmReleaseBuilder) WithLabels(labels map[string]string) *HelmReleaseBuilder {
	for k, v := range labels {
		b.labels[k] = v
	}
	return b
}

// WithAnnotation adds an annotation to the HelmRelease
func (b *HelmReleaseBuilder) WithAnnotation(key, value string) *HelmReleaseBuilder {
	b.annotations[key] = value
	return b
}

// Build creates the typed HelmRelease object
func (b *HelmReleaseBuilder) Build() *helmv1.Release {
	// Marshal values to RawExtension
	valuesRaw := runtime.RawExtension{}
	if len(b.values) > 0 {
		valuesJSON, _ := json.Marshal(b.values)
		valuesRaw.Raw = valuesJSON
	}

	return &helmv1.Release{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "helm.crossplane.io/v1beta1",
			Kind:       "Release",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        b.name,
			Labels:      b.labels,
			Annotations: b.annotations,
		},
		Spec: helmv1.ReleaseSpec{
			ForProvider: helmv1.ReleaseParameters{
				Chart: helmv1.ChartSpec{
					Repository: b.chartRepo,
					Name:       b.chartName,
					Version:    b.chartVersion,
				},
				Namespace: b.targetNamespace,
				ValuesSpec: helmv1.ValuesSpec{
					Values: valuesRaw,
				},
			},
		},
	}
}