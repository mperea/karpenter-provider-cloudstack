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

package cloudprovider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/awslabs/operatorpkg/status"
	"github.com/samber/lo"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	karpv1 "sigs.k8s.io/karpenter/pkg/apis/v1"
	"sigs.k8s.io/karpenter/pkg/cloudprovider"
	"sigs.k8s.io/karpenter/pkg/events"

	v1 "github.com/mperea/karpenter-provider-cloudstack/pkg/apis/v1"
	"github.com/mperea/karpenter-provider-cloudstack/pkg/providers/instance"
	"github.com/mperea/karpenter-provider-cloudstack/pkg/providers/instancetype"
)

var _ cloudprovider.CloudProvider = (*CloudProvider)(nil)

// CloudProvider implements the Karpenter CloudProvider interface for CloudStack
type CloudProvider struct {
	kubeClient           client.Client
	recorder             events.Recorder
	instanceTypeProvider instancetype.Provider
	instanceProvider     instance.Provider
}

// New creates a new CloudStack cloud provider
func New(
	instanceTypeProvider instancetype.Provider,
	instanceProvider instance.Provider,
	recorder events.Recorder,
	kubeClient client.Client,
) *CloudProvider {
	return &CloudProvider{
		instanceTypeProvider: instanceTypeProvider,
		instanceProvider:     instanceProvider,
		kubeClient:           kubeClient,
		recorder:             recorder,
	}
}

// Create provisions a new node in CloudStack
func (c *CloudProvider) Create(ctx context.Context, nodeClaim *karpv1.NodeClaim) (*karpv1.NodeClaim, error) {
	nodeClass, err := c.resolveNodeClassFromNodeClaim(ctx, nodeClaim)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, cloudprovider.NewInsufficientCapacityError(fmt.Errorf("resolving nodeclass: %w", err))
		}
		return nil, fmt.Errorf("resolving nodeclass: %w", err)
	}

	// Check if NodeClass is ready
	readyCondition := nodeClass.GetCondition("Ready")
	if readyCondition == nil || readyCondition.Status != metav1.ConditionTrue {
		return nil, cloudprovider.NewNodeClassNotReadyError(fmt.Errorf("nodeclass %s is not ready", nodeClass.Name))
	}

	// Get instance types
	instanceTypes, err := c.instanceTypeProvider.List(ctx, nodeClass)
	if err != nil {
		return nil, cloudprovider.NewCreateError(fmt.Errorf("resolving instance types: %w", err), "InstanceTypeResolutionFailed", "Error resolving instance types")
	}

	// Create the instance
	inst, err := c.instanceProvider.Create(ctx, nodeClass, nodeClaim, instanceTypes)
	if err != nil {
		return nil, fmt.Errorf("creating instance: %w", err)
	}

	// Convert instance to NodeClaim
	nc := c.instanceToNodeClaim(inst, nodeClass)
	nc.Annotations = lo.Assign(nc.Annotations, map[string]string{
		v1.AnnotationNodeClassHash:        nodeClass.Hash(),
		v1.AnnotationNodeClassHashVersion: v1.CloudStackNodeClassHashVersion,
	})

	log.FromContext(ctx).Info("Created node", "nodeClaim", nodeClaim.Name, "instanceID", inst.ID)

	return nc, nil
}

// Get retrieves a node by provider ID
func (c *CloudProvider) Get(ctx context.Context, providerID string) (*karpv1.NodeClaim, error) {
	id, err := ParseProviderID(providerID)
	if err != nil {
		return nil, fmt.Errorf("parsing provider ID: %w", err)
	}

	inst, err := c.instanceProvider.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getting instance: %w", err)
	}

	// Try to resolve node class from instance tags
	nodeClass, err := c.resolveNodeClassFromInstance(ctx, inst)
	if client.IgnoreNotFound(err) != nil {
		return nil, fmt.Errorf("resolving nodeclass: %w", err)
	}

	return c.instanceToNodeClaim(inst, nodeClass), nil
}

// List returns all nodes managed by Karpenter
func (c *CloudProvider) List(ctx context.Context) ([]*karpv1.NodeClaim, error) {
	instances, err := c.instanceProvider.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing instances: %w", err)
	}

	nodeClaims := make([]*karpv1.NodeClaim, 0, len(instances))
	for _, inst := range instances {
		nodeClass, err := c.resolveNodeClassFromInstance(ctx, inst)
		if client.IgnoreNotFound(err) != nil {
			return nil, fmt.Errorf("resolving nodeclass: %w", err)
		}

		nodeClaims = append(nodeClaims, c.instanceToNodeClaim(inst, nodeClass))
	}

	return nodeClaims, nil
}

// Delete terminates a node
func (c *CloudProvider) Delete(ctx context.Context, nodeClaim *karpv1.NodeClaim) error {
	id, err := ParseProviderID(nodeClaim.Status.ProviderID)
	if err != nil {
		return fmt.Errorf("parsing provider ID: %w", err)
	}

	if err := c.instanceProvider.Delete(ctx, id); err != nil {
		return fmt.Errorf("deleting instance: %w", err)
	}

	log.FromContext(ctx).Info("Deleted node", "nodeClaim", nodeClaim.Name, "instanceID", id)

	return nil
}

// GetInstanceTypes returns available instance types for a NodePool
func (c *CloudProvider) GetInstanceTypes(ctx context.Context, nodePool *karpv1.NodePool) ([]*cloudprovider.InstanceType, error) {
	nodeClass, err := c.resolveNodeClassFromNodePool(ctx, nodePool)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("resolving nodeclass: %w", err)
	}

	instanceTypes, err := c.instanceTypeProvider.List(ctx, nodeClass)
	if err != nil {
		return nil, err
	}

	return instanceTypes, nil
}

// IsDrifted checks if a node has drifted from its desired state
func (c *CloudProvider) IsDrifted(ctx context.Context, nodeClaim *karpv1.NodeClaim) (cloudprovider.DriftReason, error) {
	// Get NodePool name
	nodePoolName, ok := nodeClaim.Labels[karpv1.NodePoolLabelKey]
	if !ok {
		return "", nil
	}

	// Get NodePool
	nodePool := &karpv1.NodePool{}
	if err := c.kubeClient.Get(ctx, types.NamespacedName{Name: nodePoolName}, nodePool); err != nil {
		return "", client.IgnoreNotFound(err)
	}

	if nodePool.Spec.Template.Spec.NodeClassRef == nil {
		return "", nil
	}

	// Get NodeClass
	nodeClass, err := c.resolveNodeClassFromNodePool(ctx, nodePool)
	if err != nil {
		if errors.IsNotFound(err) {
			return "", nil
		}
		return "", fmt.Errorf("resolving nodeclass: %w", err)
	}

	// Check if hash has changed
	currentHash := nodeClaim.Annotations[v1.AnnotationNodeClassHash]
	expectedHash := nodeClass.Hash()

	if currentHash != expectedHash {
		return "NodeClassDrifted", nil
	}

	return "", nil
}

// Name returns the cloud provider name
func (c *CloudProvider) Name() string {
	return "cloudstack"
}

// GetSupportedNodeClasses returns the supported node class types
func (c *CloudProvider) GetSupportedNodeClasses() []status.Object {
	return []status.Object{&v1.CloudStackNodeClass{}}
}

// RepairPolicies returns the repair policies
func (c *CloudProvider) RepairPolicies() []cloudprovider.RepairPolicy {
	return []cloudprovider.RepairPolicy{
		{
			ConditionType:      corev1.NodeReady,
			ConditionStatus:    corev1.ConditionFalse,
			TolerationDuration: 30 * time.Minute,
		},
		{
			ConditionType:      corev1.NodeReady,
			ConditionStatus:    corev1.ConditionUnknown,
			TolerationDuration: 30 * time.Minute,
		},
	}
}

// DisruptionReasons returns reasons for disruption
func (c *CloudProvider) DisruptionReasons() []karpv1.DisruptionReason {
	return nil
}

// resolveNodeClassFromNodeClaim resolves the NodeClass from a NodeClaim
func (c *CloudProvider) resolveNodeClassFromNodeClaim(ctx context.Context, nodeClaim *karpv1.NodeClaim) (*v1.CloudStackNodeClass, error) {
	nodeClass := &v1.CloudStackNodeClass{}
	if err := c.kubeClient.Get(ctx, types.NamespacedName{Name: nodeClaim.Spec.NodeClassRef.Name}, nodeClass); err != nil {
		return nil, err
	}

	if !nodeClass.DeletionTimestamp.IsZero() {
		return nil, errors.NewNotFound(v1.SchemeGroupVersion.WithResource("cloudstacknodeclasses").GroupResource(), nodeClass.Name)
	}

	return nodeClass, nil
}

// resolveNodeClassFromNodePool resolves the NodeClass from a NodePool
func (c *CloudProvider) resolveNodeClassFromNodePool(ctx context.Context, nodePool *karpv1.NodePool) (*v1.CloudStackNodeClass, error) {
	nodeClass := &v1.CloudStackNodeClass{}
	if err := c.kubeClient.Get(ctx, types.NamespacedName{Name: nodePool.Spec.Template.Spec.NodeClassRef.Name}, nodeClass); err != nil {
		return nil, err
	}

	if !nodeClass.DeletionTimestamp.IsZero() {
		return nil, errors.NewNotFound(v1.SchemeGroupVersion.WithResource("cloudstacknodeclasses").GroupResource(), nodeClass.Name)
	}

	return nodeClass, nil
}

// resolveNodeClassFromInstance resolves the NodeClass from an instance
func (c *CloudProvider) resolveNodeClassFromInstance(ctx context.Context, inst *instance.Instance) (*v1.CloudStackNodeClass, error) {
	nodeClassName, ok := inst.Tags[v1.NodeClassTagKey]
	if !ok {
		return nil, errors.NewNotFound(v1.SchemeGroupVersion.WithResource("cloudstacknodeclasses").GroupResource(), "")
	}

	nodeClass := &v1.CloudStackNodeClass{}
	if err := c.kubeClient.Get(ctx, types.NamespacedName{Name: nodeClassName}, nodeClass); err != nil {
		return nil, err
	}

	if !nodeClass.DeletionTimestamp.IsZero() {
		return nil, errors.NewNotFound(v1.SchemeGroupVersion.WithResource("cloudstacknodeclasses").GroupResource(), nodeClass.Name)
	}

	return nodeClass, nil
}

// instanceToNodeClaim converts an Instance to a NodeClaim
func (c *CloudProvider) instanceToNodeClaim(inst *instance.Instance, nodeClass *v1.CloudStackNodeClass) *karpv1.NodeClaim {
	nodeClaim := &karpv1.NodeClaim{}

	labels := map[string]string{
		corev1.LabelTopologyZone:       inst.Zone,
		corev1.LabelInstanceTypeStable: inst.ServiceOffering,
		v1.LabelCapacityType:           v1.CapacityTypeOnDemand,
		corev1.LabelArchStable:         v1.ArchitectureAmd64, // Default
		corev1.LabelOSStable:           v1.OSLinux,
		v1.LabelZoneID:                 inst.ZoneID,
		v1.LabelZoneName:               inst.Zone,
		v1.LabelNetworkID:              inst.NetworkID,
		v1.LabelServiceOfferingID:      inst.ServiceOfferingID,
		v1.LabelServiceOfferingName:    inst.ServiceOffering,
		v1.LabelTemplateID:             inst.TemplateID,
		v1.LabelTemplateName:           inst.Template,
	}

	// Add NodePool label if present
	if nodePoolName, ok := inst.Tags[v1.NodePoolTagKey]; ok {
		labels[karpv1.NodePoolLabelKey] = nodePoolName
	}

	nodeClaim.Labels = labels
	nodeClaim.Annotations = map[string]string{}
	nodeClaim.CreationTimestamp = metav1.Time{Time: inst.CreatedTime}

	// Set deletion timestamp if VM is being terminated
	if inst.State == "Destroyed" || inst.State == "Expunging" {
		now := metav1.Now()
		nodeClaim.DeletionTimestamp = &now
	}

	// Set provider ID
	nodeClaim.Status.ProviderID = FormatProviderID(inst.Zone, inst.ID)

	// Set image ID
	nodeClaim.Status.ImageID = inst.TemplateID

	// TODO: Set capacity and allocatable based on service offering
	// This would require fetching the service offering details

	return nodeClaim
}

// ParseProviderID parses a provider ID and returns the instance ID
func ParseProviderID(providerID string) (string, error) {
	// Format: cloudstack://<zone>/<instance-id>
	// For simplicity, we'll extract just the instance ID
	if providerID == "" {
		return "", fmt.Errorf("provider ID is empty")
	}

	// Simple parsing - extract the last part after the last /
	parts := strings.Split(providerID, "/")
	if len(parts) == 0 {
		return "", fmt.Errorf("invalid provider ID format: %s", providerID)
	}
	parts = lo.Reverse(parts)

	return parts[0], nil
}

// FormatProviderID formats a provider ID
func FormatProviderID(zone, instanceID string) string {
	return fmt.Sprintf("cloudstack://%s/%s", zone, instanceID)
}
