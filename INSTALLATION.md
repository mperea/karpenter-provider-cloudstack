# Installation Guide

This guide will walk you through installing the Karpenter CloudStack Provider in your Kubernetes cluster.

## Prerequisites

- Kubernetes cluster version 1.28 or later
- CloudStack 4.22 or later
- kubectl configured to access your cluster
- Helm 3.x installed
- CloudStack API credentials with appropriate permissions

## CloudStack Permissions

The CloudStack API credentials need the following permissions:

- Virtual Machine: Deploy, List, Destroy
- Service Offering: List
- Template: List
- Network: List
- Zone: List
- Tags: Create, List, Delete

## Installation Steps

### 1. Add Helm Repository

```bash
helm repo add karpenter-cloudstack https://mperea.github.io/karpenter-provider-cloudstack
helm repo update
```

### 2. Create Namespace

```bash
kubectl create namespace karpenter
```

### 3. Install Karpenter CloudStack Provider

```bash
helm install karpenter-cloudstack karpenter-cloudstack/karpenter-cloudstack \
  --namespace karpenter \
  --set cloudstack.apiUrl=https://your-cloudstack-api.com \
  --set cloudstack.apiKey=YOUR_API_KEY \
  --set cloudstack.secretKey=YOUR_SECRET_KEY \
  --set clusterName=my-cluster
```

### 4. Verify Installation

```bash
kubectl get pods -n karpenter
kubectl logs -n karpenter -l app.kubernetes.io/name=karpenter-cloudstack
```

You should see the controller running and logs indicating it has connected to CloudStack.

## Configuration Options

### values.yaml Configuration

Create a `values.yaml` file with your configuration:

```yaml
# CloudStack Configuration
cloudstack:
  apiUrl: "https://your-cloudstack-api.com"
  apiKey: "your-api-key"
  apiKey: "your-secret-key"
  verifySSL: true

# Cluster Configuration
clusterName: "my-cluster"

# Controller Resources
resources:
  limits:
    cpu: 1000m
    memory: 512Mi
  requests:
    cpu: 100m
    memory: 256Mi

# Log Level
logLevel: info
```

Install with custom values:

```bash
helm install karpenter-cloudstack karpenter-cloudstack/karpenter-cloudstack \
  --namespace karpenter \
  --values values.yaml
```

## Creating Your First NodeClass

After installation, create a CloudStackNodeClass:

```bash
kubectl apply -f - <<EOF
apiVersion: karpenter.k8s.cloudstack/v1
kind: CloudStackNodeClass
metadata:
  name: default
spec:
  zone: zone-01
  networkSelectorTerms:
    - tags:
        karpenter.sh/discovery: my-cluster
  serviceOfferingSelectorTerms:
    - tags:
        karpenter.sh/discovery: my-cluster
  templateSelectorTerms:
    - tags:
        os: ubuntu
        version: "22.04"
  userData: |
    #!/bin/bash
    echo "Karpenter node initialization"
EOF
```

## Creating Your First NodePool

Create a NodePool that references the NodeClass:

```bash
kubectl apply -f - <<EOF
apiVersion: karpenter.sh/v1
kind: NodePool
metadata:
  name: default
spec:
  template:
    spec:
      nodeClassRef:
        group: karpenter.k8s.cloudstack
        kind: CloudStackNodeClass
        name: default
      requirements:
        - key: karpenter.sh/capacity-type
          operator: In
          values: ["on-demand"]
        - key: node.kubernetes.io/instance-type
          operator: In
          values: ["Medium Instance", "Large Instance"]
  limits:
    cpu: "100"
    memory: 100Gi
EOF
```

## Testing

Deploy a test workload to verify Karpenter is working:

```bash
kubectl apply -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: inflate
spec:
  replicas: 5
  selector:
    matchLabels:
      app: inflate
  template:
    metadata:
      labels:
        app: inflate
    spec:
      containers:
      - name: inflate
        image: public.ecr.aws/eks-distro/kubernetes/pause:3.7
        resources:
          requests:
            cpu: 1
            memory: 2Gi
EOF
```

Watch for new nodes being created:

```bash
kubectl get nodes --watch
kubectl get nodeclaims --watch
```

## Tagging CloudStack Resources

For Karpenter to discover CloudStack resources, you need to tag them appropriately:

### Networks

Tag your networks with:
```
karpenter.sh/discovery: my-cluster
```

### Service Offerings

Tag your service offerings with:
```
karpenter.sh/discovery: my-cluster
```

### Templates

Tag your templates with appropriate metadata:
```
os: ubuntu
version: 22.04
karpenter.sh/discovery: my-cluster
```

## Troubleshooting

### Controller Not Starting

Check logs:
```bash
kubectl logs -n karpenter -l app.kubernetes.io/name=karpenter-cloudstack
```

Common issues:
- Incorrect CloudStack API credentials
- Network connectivity to CloudStack API
- SSL certificate verification issues

### Nodes Not Being Created

Check NodeClass status:
```bash
kubectl get cloudstacknodeclass default -o yaml
```

Check NodeClaim events:
```bash
kubectl get events --sort-by='.lastTimestamp' | grep NodeClaim
```

### SSL Certificate Issues

If you have self-signed certificates:
```bash
helm upgrade karpenter-cloudstack karpenter-cloudstack/karpenter-cloudstack \
  --namespace karpenter \
  --set cloudstack.verifySSL=false \
  --reuse-values
```

## Upgrading

To upgrade to a new version:

```bash
helm repo update
helm upgrade karpenter-cloudstack karpenter-cloudstack/karpenter-cloudstack \
  --namespace karpenter \
  --reuse-values
```

## Uninstallation

To remove Karpenter CloudStack Provider:

```bash
# Delete all NodePools and NodeClasses first
kubectl delete nodepools --all
kubectl delete cloudstacknodeclasses --all

# Uninstall Helm chart
helm uninstall karpenter-cloudstack --namespace karpenter

# Delete namespace
kubectl delete namespace karpenter
```

## Next Steps

- Review the [examples](./examples/) directory for advanced configurations
- Read the [CloudStackNodeClass specification](./docs/nodeclass-spec.md)
- Check out [best practices](./docs/best-practices.md)
- Join the community discussions

