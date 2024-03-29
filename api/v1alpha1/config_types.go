package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories=common;config;giantswarm
// +k8s:openapi-gen=true

// Config represents configuration of an App.
type Config struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ConfigSpec `json:"spec"`
	// +kubebuilder:validation:Optional
	// Status part of the Config resource.
	Status ConfigStatus `json:"status,omitempty"`
}

// ConfigSpec is the spec part for the Config resource.
// +k8s:openapi-gen=true
type ConfigSpec struct {
	// App details for which the configuration should be generated.
	App ConfigSpecApp `json:"app"`
}

// ConfigSpecApp holds the information about the App to be configured.
// +k8s:openapi-gen=true
type ConfigSpecApp struct {
	// Name is the name of the App.
	Name string `json:"name"`
	// Version is the version of the App.
	Version string `json:"version"`
	// Catalog is the name of the App's App Catalog.
	Catalog string `json:"catalog"`
}

// ConfigStatus holds status information about the generated configuration.
// +k8s:openapi-gen=true
type ConfigStatus struct {
	// +kubebuilder:validation:Optional
	// App details for which the configuration was generated.
	App ConfigStatusApp `json:"app,omitempty"`
	// +kubebuilder:validation:Optional
	// Config holds the references to the generated configuration.
	Config ConfigStatusConfig `json:"config,omitempty"`
	// +kubebuilder:validation:Optional
	// Version of the giantswarm/config repository used to generate the
	// configuration.
	Version string `json:"version,omitempty"`
}

// ConfigStatusApp holds the information about the App used to generate
// referenced configuration.
// +k8s:openapi-gen=true
type ConfigStatusApp struct {
	// Name is the name of the App.
	Name string `json:"name"`
	// Version is the version of the App.
	Version string `json:"version"`
	// Catalog is the name of the App's App Catalog.
	Catalog string `json:"catalog"`
}

// ConfigStatusConfig holds configuration ConfigMap and Secret references to be
// used to configure the App.
// +k8s:openapi-gen=true
type ConfigStatusConfig struct {
	ConfigMapRef ConfigStatusConfigConfigMapRef `json:"configMapRef"`
	SecretRef    ConfigStatusConfigSecretRef    `json:"secretRef"`
}

// ConfigStatusConfigConfigMapRef contains a reference to the generated ConfigMap.
// +k8s:openapi-gen=true
type ConfigStatusConfigConfigMapRef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// ConfigStatusConfigSecretRef contains a reference to the generated Secret.
// +k8s:openapi-gen=true
type ConfigStatusConfigSecretRef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Config `json:"items"`
}
