/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

const (
	// ClusterFinalizer allows ReconcileHetznerCluster to clean up HCloud
	// resources associated with HetznerCluster before removing it from the
	// apiserver.
	ClusterFinalizer = "hetznercluster.infrastructure.cluster.x-k8s.io"

	// HCloudLoadBalancerAlgorithmTypeRoundRobin default for the Kubernetes Api Server loadbalancer.
	HCloudLoadBalancerAlgorithmTypeRoundRobin = HCloudLoadBalancerAlgorithmType("round_robin")

	// HCloudLoadBalancerAlgorithmTypeLeastConnections default for Loadbalancer.
	HCloudLoadBalancerAlgorithmTypeLeastConnections = HCloudLoadBalancerAlgorithmType("least_connections")
)

// HetznerClusterSpec defines the desired state of HetznerCluster.
type HetznerClusterSpec struct {
	// NetworkSpec defines the Network for Hetzner Cloud. If left empty no private Network is configured.
	// +optional
	NetworkSpec NetworkSpec `json:"network"`

	// Hetzner Regions (fsn, nbg, hel). Because Hetzner Networks have a very low latency we could assume in some use-cases
	// that a region is behaving like a zone https://kubernetes.io/docs/reference/labels-annotations-taints/#topologykubernetesiozone
	// therefore this accepts a list of the hetzner regions.
	Region []HCloudRegion `json:"controlPlaneRegion"`

	// define cluster wide SSH keys. Valid values are a valid SSH key name, or a valid ID.
	SSHKey []HCloudSSHKeySpec `json:"sshKey,omitempty"`
	// ControlPlaneEndpoint represents the endpoint used to communicate with the control plane.
	// +optional
	ControlPlaneEndpoint *clusterv1.APIEndpoint `json:"controlPlaneEndpoint"`

	// ControlPlaneLoadBalancer is optional configuration for customizing control plane behavior. Naming convention is from upstream cluster-api project.
	// +optional
	ControlPlaneLoadBalancer HCloudLoadBalancerSpec `json:"controlPlaneLoadBalancer,omitempty"`

	// +optional
	PlacementGroupSpec []HCloudPlacementGroupSpec `json:"placementGroup,omitempty"`

	// HcloudTokenRef is a reference to a token to be used when reconciling this cluster
	// this is generated in the Security section under API TOKENS. Read & Write is necessary
	HCloudTokenRef *corev1.SecretKeySelector `json:"hcloudTokenRef"`
}

// HCloudLoadBalancerSpec defines the desired state of the Control Plane Loadbalancer.
type HCloudLoadBalancerSpec struct {
	// +optional
	Name *string `json:"name,omitempty"`

	// Could be round-robin or least-connection
	Algorithm HCloudLoadBalancerAlgorithmType `json:"algorithm"`

	// Loadbalancer type
	// +kubebuilder:validation:Enum=lb11;lb21;lb31
	Type string `json:"type"`

	// Defines how traffic will be routed from the Load Balancer to your target server.
	Services []LoadBalancerServiceSpec `json:"services"`

	// +kubebuilder:validation:Enum=fsn1;hel1;nbg1
	Region string `json:"region"`

	ID string `json:"id,omitempty"`
}

// HCloudLoadBalancerStatus defines the obeserved state of the control plane loadbalancer.
type HCloudLoadBalancerStatus struct {
	ID                int                             `json:"id,omitempty"`
	Name              string                          `json:"name,omitempty"`
	Type              string                          `json:"type,omitempty"`
	IPv4              string                          `json:"ipv4,omitempty"`
	IPv6              string                          `json:"ipv6,omitempty"`
	InternalIP        string                          `json:"internalIP,omitempty"`
	Labels            map[string]string               `json:"labels,omitempty"`
	Algorithm         HCloudLoadBalancerAlgorithmType `json:"algorithm,omitempty"`
	Target            []int                           `json:"target,omitempty"`
	AttachedToNetwork bool                            `json:"attachedToNetwork,omitempty"`
}

// HetznerClusterStatus defines the observed state of HetznerCluster.
type HetznerClusterStatus struct {
	// +kubebuilder:default=false
	Ready bool `json:"ready"`

	// +optional
	Network *NetworkStatus `json:"networkStatus,omitempty"`

	Region                   []HCloudRegion           `json:"controlPlaneRegion,omitempty"`
	ControlPlaneLoadBalancer HCloudLoadBalancerStatus `json:"controlPlaneLoadBalancer,omitempty"`
	// +optional
	PlacementGroup []HCloudPlacementGroupStatus `json:"placementGroup,omitempty"`
	FailureDomains clusterv1.FailureDomains     `json:"failureDomains,omitempty"`
	Conditions     clusterv1.Conditions         `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=hetznerclusters,scope=Namespaced,categories=cluster-api,shortName=capihc
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Cluster",type="string",JSONPath=".metadata.labels.cluster\\.x-k8s\\.io/cluster-name",description="Cluster to which this HetznerCluster belongs"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.ready",description="Cluster infrastructure is ready for Nodes"
// +kubebuilder:printcolumn:name="Endpoint",type="string",JSONPath=".spec.controlPlaneEndpoint",description="API Endpoint",priority=1
// +k8s:defaulter-gen=true

// HetznerCluster is the Schema for the hetznercluster API.
type HetznerCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HetznerClusterSpec   `json:"spec,omitempty"`
	Status HetznerClusterStatus `json:"status,omitempty"`
}

// HetznerClusterList contains a list of HetznerCluster
// +kubebuilder:object:root=true
// +k8s:defaulter-gen=true
type HetznerClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HetznerCluster `json:"items"`
}

// GetConditions returns the observations of the operational state of the HetznerCluster resource.
func (r *HetznerCluster) GetConditions() clusterv1.Conditions {
	return r.Status.Conditions
}

// SetConditions sets the underlying service state of the HetznerCluster to the predescribed clusterv1.Conditions.
func (r *HetznerCluster) SetConditions(conditions clusterv1.Conditions) {
	r.Status.Conditions = conditions
}

func init() {
	SchemeBuilder.Register(&HetznerCluster{}, &HetznerClusterList{})
}