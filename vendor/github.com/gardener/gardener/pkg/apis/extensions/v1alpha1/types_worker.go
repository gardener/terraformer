// Copyright (c) 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var _ Object = (*Worker)(nil)

// WorkerResource is a constant for the name of the Worker resource.
const WorkerResource = "Worker"

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:scope=Namespaced,path=workers,singular=worker
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name=Type,JSONPath=".spec.type",type=string,description="The type of the cloud provider for this resource."
// +kubebuilder:printcolumn:name=Region,JSONPath=".spec.region",type=string,description="The region into which the worker should be deployed."
// +kubebuilder:printcolumn:name=Status,JSONPath=".status.lastOperation.state",type=string,description="Status of the worker."
// +kubebuilder:printcolumn:name=Age,JSONPath=".metadata.creationTimestamp",type=date,description="creation timestamp"

// Worker is a specification for a Worker resource.
type Worker struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Specification of the Worker.
	// If the object's deletion timestamp is set, this field is immutable.
	Spec WorkerSpec `json:"spec"`
	// +optional
	Status WorkerStatus `json:"status"`
}

// GetExtensionSpec implements Object.
func (i *Worker) GetExtensionSpec() Spec {
	return &i.Spec
}

// GetExtensionStatus implements Object.
func (i *Worker) GetExtensionStatus() Status {
	return &i.Status
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WorkerList is a list of Worker resources.
type WorkerList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	// Items is the list of Worker.
	Items []Worker `json:"items"`
}

// WorkerSpec is the spec for a Worker resource.
type WorkerSpec struct {
	// DefaultSpec is a structure containing common fields used by all extension resources.
	DefaultSpec `json:",inline"`

	// InfrastructureProviderStatus is a raw extension field that contains the provider status that has
	// been generated by the controller responsible for the `Infrastructure` resource.
	// +kubebuilder:validation:XPreserveUnknownFields
	// +kubebuilder:pruning:PreserveUnknownFields
	// +optional
	InfrastructureProviderStatus *runtime.RawExtension `json:"infrastructureProviderStatus,omitempty"`
	// Region is the name of the region where the worker pool should be deployed to. This field is immutable.
	Region string `json:"region"`
	// SecretRef is a reference to a secret that contains the cloud provider specific credentials.
	SecretRef corev1.SecretReference `json:"secretRef"`
	// SSHPublicKey is the public SSH key that should be used with these workers.
	// +optional
	SSHPublicKey []byte `json:"sshPublicKey,omitempty"`
	// Pools is a list of worker pools.
	// +patchMergeKey=name
	// +patchStrategy=merge
	Pools []WorkerPool `json:"pools" patchStrategy:"merge" patchMergeKey:"name"`
}

// WorkerPool is the definition of a specific worker pool.
type WorkerPool struct {
	// MachineType contains information about the machine type that should be used for this worker pool.
	MachineType string `json:"machineType"`
	// Maximum is the maximum size of the worker pool.
	Maximum int32 `json:"maximum"`
	// MaxSurge is maximum number of VMs that are created during an update.
	MaxSurge intstr.IntOrString `json:"maxSurge"`
	// MaxUnavailable is the maximum number of VMs that can be unavailable during an update.
	MaxUnavailable intstr.IntOrString `json:"maxUnavailable"`
	// Annotations is a map of key/value pairs for annotations for all the `Node` objects in this worker pool.
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
	// Labels is a map of key/value pairs for labels for all the `Node` objects in this worker pool.
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
	// Taints is a list of taints for all the `Node` objects in this worker pool.
	// +optional
	Taints []corev1.Taint `json:"taints,omitempty"`
	// MachineImage contains logical information about the name and the version of the machie image that
	// should be used. The logical information must be mapped to the provider-specific information (e.g.,
	// AMIs, ...) by the provider itself.
	MachineImage MachineImage `json:"machineImage,omitempty"`
	// Minimum is the minimum size of the worker pool.
	Minimum int32 `json:"minimum"`
	// Name is the name of this worker pool.
	Name string `json:"name"`
	// ProviderConfig is a provider specific configuration for the worker pool.
	// +kubebuilder:validation:XPreserveUnknownFields
	// +kubebuilder:pruning:PreserveUnknownFields
	// +optional
	ProviderConfig *runtime.RawExtension `json:"providerConfig,omitempty"`
	// UserData is a base64-encoded string that contains the data that is sent to the provider's APIs
	// when a new machine/VM that is part of this worker pool shall be spawned.
	UserData []byte `json:"userData"`
	// Volume contains information about the root disks that should be used for this worker pool.
	// +optional
	Volume *Volume `json:"volume,omitempty"`
	// DataVolumes contains a list of additional worker volumes.
	// +optional
	DataVolumes []DataVolume `json:"dataVolumes,omitempty"`
	// KubeletDataVolumeName contains the name of a dataVolume that should be used for storing kubelet state.
	// +optional
	KubeletDataVolumeName *string `json:"kubeletDataVolumeName,omitempty"`
	// Zones contains information about availability zones for this worker pool.
	// +optional
	Zones []string `json:"zones,omitempty"`
	// MachineControllerManagerSettings contains configurations for different worker-pools. Eg. MachineDrainTimeout, MachineHealthTimeout.
	// +optional
	MachineControllerManagerSettings *gardencorev1beta1.MachineControllerManagerSettings `json:"machineControllerManager,omitempty"`
	// KubernetesVersion is the kubernetes version in this worker pool
	// +optional
	KubernetesVersion *string `json:"kubernetesVersion,omitempty"`
	// NodeTemplate contains resource information of the machine which is used by Cluster Autoscaler to generate nodeTemplate during scaling a nodeGroup from zero
	// +optional
	NodeTemplate *NodeTemplate `json:"nodeTemplate,omitempty"`
	// Architecture is the CPU architecture of the worker pool machines and machine image.
	// +optional
	Architecture *string `json:"architecture,omitempty"`
}

// NodeTemplate contains information about the expected node properties.
type NodeTemplate struct {
	// Capacity represents the expected Node capacity.
	Capacity corev1.ResourceList `json:"capacity"`
}

// MachineImage contains logical information about the name and the version of the machie image that
// should be used. The logical information must be mapped to the provider-specific information (e.g.,
// AMIs, ...) by the provider itself.
type MachineImage struct {
	// Name is the logical name of the machine image.
	Name string `json:"name"`
	// Version is the version of the machine image.
	Version string `json:"version"`
}

// Volume contains information about the root disks that should be used for worker pools.
type Volume struct {
	// Name of the volume to make it referencable.
	// +optional
	Name *string `json:"name,omitempty"`
	// Type is the type of the volume.
	// +optional
	Type *string `json:"type,omitempty"`
	// Size is the of the root volume.
	Size string `json:"size"`
	// Encrypted determines if the volume should be encrypted.
	// +optional
	Encrypted *bool `json:"encrypted,omitempty"`
}

// DataVolume contains information about a data volume.
type DataVolume struct {
	// Name of the volume to make it referencable.
	Name string `json:"name"`
	// Type is the type of the volume.
	// +optional
	Type *string `json:"type,omitempty"`
	// Size is the of the root volume.
	Size string `json:"size"`
	// Encrypted determines if the volume should be encrypted.
	// +optional
	Encrypted *bool `json:"encrypted,omitempty"`
}

// WorkerStatus is the status for a Worker resource.
type WorkerStatus struct {
	// DefaultStatus is a structure containing common fields used by all extension resources.
	DefaultStatus `json:",inline"`
	// MachineDeployments is a list of created machine deployments. It will be used to e.g. configure
	// the cluster-autoscaler properly.
	// +patchMergeKey=name
	// +patchStrategy=merge
	MachineDeployments []MachineDeployment `json:"machineDeployments,omitempty" patchStrategy:"merge" patchMergeKey:"name"`
}

// MachineDeployment is a created machine deployment.
type MachineDeployment struct {
	// Name is the name of the `MachineDeployment` resource.
	Name string `json:"name"`
	// Minimum is the minimum number for this machine deployment.
	Minimum int32 `json:"minimum"`
	// Maximum is the maximum number for this machine deployment.
	Maximum int32 `json:"maximum"`
}
