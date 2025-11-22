/*
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

package v1

import (
	corev1 "k8s.io/api/core/v1"
)

const (
	// CloudStack-specific labels
	LabelTopologyZone   = corev1.LabelTopologyZone   // topology.kubernetes.io/zone
	LabelTopologyRegion = corev1.LabelTopologyRegion // topology.kubernetes.io/region
	LabelInstanceType   = corev1.LabelInstanceTypeStable
	LabelArchitecture   = corev1.LabelArchStable
	LabelOS             = corev1.LabelOSStable

	// CloudStack specific labels
	LabelZoneID              = "karpenter.k8s.cloudstack/zone-id"
	LabelZoneName            = "karpenter.k8s.cloudstack/zone-name"
	LabelNetworkID           = "karpenter.k8s.cloudstack/network-id"
	LabelNetworkName         = "karpenter.k8s.cloudstack/network-name"
	LabelServiceOfferingID   = "karpenter.k8s.cloudstack/service-offering-id"
	LabelServiceOfferingName = "karpenter.k8s.cloudstack/service-offering-name"
	LabelTemplateID          = "karpenter.k8s.cloudstack/template-id"
	LabelTemplateName        = "karpenter.k8s.cloudstack/template-name"
	LabelDiskOfferingID      = "karpenter.k8s.cloudstack/disk-offering-id"

	// Capacity type - CloudStack only supports on-demand
	LabelCapacityType    = "karpenter.sh/capacity-type"
	CapacityTypeOnDemand = "on-demand"

	// Tag keys used for resource identification
	NodePoolTagKey    = "karpenter.sh/nodepool"
	NodeClaimTagKey   = "karpenter.sh/nodeclaim"
	NodeClassTagKey   = "karpenter.k8s.cloudstack/nodeclass"
	ClusterNameTagKey = "kubernetes.io/cluster"
	ManagedByTagKey   = "karpenter.sh/managed-by"

	// Annotations
	AnnotationNodeClassHash        = "karpenter.k8s.cloudstack/nodeclass-hash"
	AnnotationNodeClassHashVersion = "karpenter.k8s.cloudstack/nodeclass-hash-version"
)

// Well-known label values
const (
	ArchitectureAmd64 = "amd64"
	ArchitectureArm64 = "arm64"
	OSLinux           = "linux"
)
