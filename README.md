# Karpenter Provider for Apache CloudStack

[![GitHub Stars](https://img.shields.io/github/stars/mperea/karpenter-provider-cloudstack?style=social)](https://github.com/mperea/karpenter-provider-cloudstack)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/mperea/karpenter-provider-cloudstack)](https://goreportcard.com/report/github.com/mperea/karpenter-provider-cloudstack)

A [Karpenter](https://karpenter.sh) cloud provider implementation for Apache CloudStack.

> **Author:** Mario Perea ([@mperea](https://github.com/mperea))
> This project is developed in my personal time as a contribution to the CloudStack and Kubernetes communities.

## Overview

This project provides a CloudStack-specific implementation of Karpenter's cloud provider interface, enabling automatic node provisioning and scaling for Kubernetes clusters running on CloudStack infrastructure.

## Features

- **Automatic Node Provisioning**: Dynamically provision and deprovision nodes based on pod requirements
- **Service Offering Selection**: Automatically select appropriate CloudStack service offerings based on resource requirements
- **Network Management**: Support for CloudStack network selection and configuration
- **Template Management**: Flexible template selection for different workload types
- **Zone Support**: Multi-zone deployment capabilities
- **Tag-based Resource Selection**: Use CloudStack tags to filter and select resources

## Architecture

The provider follows Karpenter's extensible architecture with CloudStack-specific implementations:

- **CloudStackNodeClass CRD**: Defines CloudStack-specific node configuration
- **Instance Provider**: Manages VM lifecycle using CloudStack API
- **InstanceType Provider**: Handles Service Offering discovery and caching
- **Network Provider**: Manages network selection and validation
- **Template Provider**: Handles template/image selection
- **Zone Provider**: Manages CloudStack zone information

## Prerequisites

- Kubernetes cluster running on CloudStack infrastructure
- CloudStack API credentials with appropriate permissions
- Karpenter core controllers installed

## Quick Start

### Installation

1. Install the Karpenter CloudStack provider:

```bash
helm repo add karpenter-cloudstack https://mperea.github.io/karpenter-provider-cloudstack
helm install karpenter-cloudstack karpenter-cloudstack/karpenter-cloudstack \
  --namespace karpenter \
  --set cloudstack.apiUrl=https://your-cloudstack-api.com \
  --set cloudstack.apiKey=YOUR_API_KEY \
  --set cloudstack.secretKey=YOUR_SECRET_KEY \
  --set clusterName=my-cluster
```

2. Create a CloudStackNodeClass:

```yaml
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
    echo "Node initialization script"
```

3. Create a NodePool:

```yaml
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
          values: ["medium", "large"]
```

## Configuration

### Environment Variables

The following environment variables are required:

| Variable | Description | Required |
|----------|-------------|----------|
| `CLOUDSTACK_API_URL` | CloudStack API endpoint URL | Yes |
| `CLOUDSTACK_API_KEY` | CloudStack API key | Yes |
| `CLOUDSTACK_SECRET_KEY` | CloudStack secret key | Yes |
| `CLOUDSTACK_VERIFY_SSL` | Verify SSL certificates (default: true) | No |
| `CLUSTER_NAME` | Kubernetes cluster name | Yes |

### CloudStackNodeClass Specification

The `CloudStackNodeClass` CRD supports the following fields:

- `zone`: CloudStack zone where VMs will be deployed
- `networkSelectorTerms`: Network selection criteria (tags, id, name)
- `serviceOfferingSelectorTerms`: Service offering selection criteria
- `templateSelectorTerms`: Template/image selection criteria
- `userData`: Cloud-init script for VM initialization
- `tags`: Tags to apply to created VMs
- `rootDiskSize`: Size of root disk (in GB)
- `diskOffering`: Disk offering for data disks

## Development

### Building

```bash
make build
```

### Testing

```bash
# Unit tests
make test

# Integration tests (requires CloudStack environment)
make test-integration
```

### Local Development

```bash
# Run locally against a Kubernetes cluster
export CLOUDSTACK_API_URL=https://your-cloudstack-api.com
export CLOUDSTACK_API_KEY=your-api-key
export CLOUDSTACK_SECRET_KEY=your-secret-key
export CLUSTER_NAME=my-cluster

go run cmd/controller/main.go
```

## Contributing

Contributions are welcome! Please read our [Contributing Guide](CONTRIBUTING.md) for details.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## References

- [Karpenter Documentation](https://karpenter.sh)
- [CloudStack API Documentation](https://cloudstack.apache.org/api/apidocs-4.22/)
- [CloudStack Go SDK](https://github.com/apache/cloudstack-go)

