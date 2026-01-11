package main

import (
	"encoding/json"

	helmv1 "github.com/crossplane-contrib/provider-helm/apis/namespaced/release/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// SecretBuilder builds Kubernetes Secret objects using fluent API
type SecretBuilder struct {
	name      string
	namespace string
	data      map[string][]byte
	labels    map[string]string
}

// NewSecretBuilder creates a new secret builder
func NewSecretBuilder(name, namespace string) *SecretBuilder {
	return &SecretBuilder{
		name:      name,
		namespace: namespace,
		data:      make(map[string][]byte),
		labels:    make(map[string]string),
	}
}

// WithData adds binary data to the secret
func (b *SecretBuilder) WithData(key string, value []byte) *SecretBuilder {
	b.data[key] = value
	return b
}

// WithLabel adds a label to the secret
func (b *SecretBuilder) WithLabel(key, value string) *SecretBuilder {
	b.labels[key] = value
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
			Name:      b.name,
			Namespace: b.namespace,
			Labels:    b.labels,
		},
		Data: b.data,
	}
}

// HelmReleaseBuilder builds helm.m.crossplane.io/v1beta1 Release objects using fluent API
type HelmReleaseBuilder struct {
	name         string
	namespace    string
	chartRepo    string
	chartName    string
	chartVersion string
	values       map[string]any
	labels       map[string]string
}

// NewHelmReleaseBuilder creates a new HelmRelease builder
func NewHelmReleaseBuilder(name string) *HelmReleaseBuilder {
	return &HelmReleaseBuilder{
		name:   name,
		values: make(map[string]any),
		labels: make(map[string]string),
	}
}

// WithNamespace sets the namespace for the HelmRelease resource
func (b *HelmReleaseBuilder) WithNamespace(namespace string) *HelmReleaseBuilder {
	b.namespace = namespace
	return b
}

// WithChart sets the Helm chart repository, name, and version
func (b *HelmReleaseBuilder) WithChart(repo, name, version string) *HelmReleaseBuilder {
	b.chartRepo = repo
	b.chartName = name
	b.chartVersion = version
	return b
}

// WithValues sets the Helm values
func (b *HelmReleaseBuilder) WithValues(values map[string]any) *HelmReleaseBuilder {
	b.values = values
	return b
}

// WithLabel adds a label to the HelmRelease
func (b *HelmReleaseBuilder) WithLabel(key, value string) *HelmReleaseBuilder {
	b.labels[key] = value
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
			APIVersion: "helm.m.crossplane.io/v1beta1",
			Kind:       "Release",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.name,
			Namespace: b.namespace,
			Labels:    b.labels,
		},
		Spec: helmv1.ReleaseSpec{
			ForProvider: helmv1.ReleaseParameters{
				Chart: helmv1.ChartSpec{
					Repository: b.chartRepo,
					Name:       b.chartName,
					Version:    b.chartVersion,
				},
				SkipCreateNamespace: true,
				ValuesSpec: helmv1.ValuesSpec{
					Values: valuesRaw,
				},
			},
		},
	}
}
