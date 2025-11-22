# Project Architecture

## Overview

Karpenter Provider for CloudStack follows the extensible Karpenter architecture, implementing CloudStack-specific interfaces.

```
┌────────────────────────────────────────────────────────┐
│                    Karpenter Core                      │
│    (Scheduling, Provisioning, Deprovisioning logic)    │
└───────────────────────┬────────────────────────────────┘
                        │ Cloud Provider Interface
┌───────────────────────▼────────────────────────────────┐
│              Karpenter CloudStack Provider             │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │  Instance    │  │ InstanceType │  │   Network    │  │
│  │  Provider    │  │  Provider    │  │   Provider   │  │
│  └──────────────┘  └──────────────┘  └──────────────┘  │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │  Template    │  │    Zone      │  │ NodeClass    │  │
│  │  Provider    │  │  Provider    │  │ Controller   │  │
│  └──────────────┘  └──────────────┘  └──────────────┘  │
└───────────────────────┬────────────────────────────────┘
                        │ CloudStack SDK
┌───────────────────────▼────────────────────────────────┐
│                  CloudStack API                        │
│   (deployVirtualMachine, listServiceOfferings, etc.)   │
└────────────────────────────────────────────────────────┘
```

---

## Main Components

### 1. Custom Resource Definitions (CRDs)

#### CloudStackNodeClass
Defines CloudStack-specific node configuration:

```yaml
apiVersion: karpenter.k8s.cloudstack/v1
kind: CloudStackNodeClass
metadata:
  name: default
spec:
  zone: zone-01
  networkSelectorTerms:
    - name: k8s-network
  serviceOfferingSelectorTerms:
    - tags:
        compute-type: general-purpose
  templateSelectorTerms:
    - name: ubuntu-22.04-k8s
  userData: |
    #!/bin/bash
    # Bootstrap script
```

**Main fields:**
- `zone`: CloudStack zone where VMs will be created
- `networkSelectorTerms`: Network selection
- `serviceOfferingSelectorTerms`: Service Offering selection
- `templateSelectorTerms`: Template/image selection
- `userData`: Initialization script
- `tags`: Tags to apply to VMs

### 2. Providers

#### Instance Provider (`pkg/providers/instance/`)
Manages VM lifecycle in CloudStack:

```go
type Provider interface {
    Create(context.Context, *v1.NodeClaim) (*v1.Instance, error)
    Get(context.Context, string) (*v1.Instance, error)
    List(context.Context) ([]*v1.Instance, error)
    Delete(context.Context, string) error
}
```

**Responsibilities:**
- Create VMs using `deployVirtualMachine`
- Query VM status
- Delete VMs using `destroyVirtualMachine`
- Map CloudStack VMs to Karpenter Instances

#### InstanceType Provider (`pkg/providers/instancetype/`)
Discovers and caches CloudStack Service Offerings:

```go
type Provider interface {
    List(context.Context, *v1.NodeClaim) ([]*cloudprovider.InstanceType, error)
    Get(context.Context, string) (*cloudprovider.InstanceType, error)
}
```

**Responsibilities:**
- List Service Offerings using `listServiceOfferings`
- Convert Service Offerings to Karpenter InstanceTypes
- Cache results for performance
- Filter by tags and requirements

#### Network Provider (`pkg/providers/network/`)
Manages network selection:

```go
type Provider interface {
    GetByName(context.Context, string, string) (*cloudstack.Network, error)
    GetByID(context.Context, string) (*cloudstack.Network, error)
    GetByTags(context.Context, string, map[string]string) ([]*cloudstack.Network, error)
}
```

**Responsibilities:**
- Search networks by name, ID, or tags
- Validate network availability
- Cache network information

#### Template Provider (`pkg/providers/template/`)
Manages template/image selection:

```go
type Provider interface {
    GetByName(context.Context, string, string) (*cloudstack.Template, error)
    GetByID(context.Context, string) (*cloudstack.Template, error)
    GetByTags(context.Context, string, map[string]string) ([]*cloudstack.Template, error)
}
```

**Responsibilities:**
- Search templates by name, ID, or tags
- Validate available templates
- Cache template information

#### Zone Provider (`pkg/providers/zone/`)
Manages CloudStack zone information:

```go
type Provider interface {
    List(context.Context) ([]*cloudstack.Zone, error)
    GetByName(context.Context, string) (*cloudstack.Zone, error)
}
```

**Responsibilities:**
- List available zones
- Validate zones
- Cache zone information

### 3. Controllers

#### NodeClass Controller (`pkg/controllers/nodeclass/`)
Reconciles CloudStackNodeClass resources:

**Responsibilities:**
- Validate CloudStackNodeClass configuration
- Verify that zone, networks, templates exist
- Update status conditions (Ready, NetworkReady, etc.)
- Detect configuration changes (drift)

### 4. Cloud Provider (`pkg/cloudprovider/`)
Implements the main Karpenter interface:

```go
type CloudProvider interface {
    Create(context.Context, *v1.NodeClaim) (*v1.NodeClaim, error)
    Get(context.Context, string) (*v1.NodeClaim, error)
    List(context.Context) ([]*v1.NodeClaim, error)
    GetInstanceTypes(context.Context, *v1.NodePool) ([]*cloudprovider.InstanceType, error)
    Delete(context.Context, *v1.NodeClaim) error
    // ...
}
```

**Responsibilities:**
- Orchestrate different providers
- Implement node creation/deletion logic
- Manage mapping between Karpenter and CloudStack
- Generate unique ProviderID for each node

---

## Node Provisioning Flow

```
1. Unscheduled Pod
   │
   ▼
2. Karpenter Core detects need
   │
   ▼
3. Calculate requirements (CPU, RAM, labels, taints)
   │
   ▼
4. CloudProvider.GetInstanceTypes()
   │
   ├─► InstanceType Provider lists Service Offerings
   │   └─► Filter by requirements
   │
   ▼
5. Select optimal InstanceType
   │
   ▼
6. CloudProvider.Create(NodeClaim)
   │
   ├─► Template Provider searches for image
   ├─► Network Provider searches for network
   └─► Instance Provider creates VM
       │
       └─► cloudstack.deployVirtualMachine()
   │
   ▼
7. VM created in CloudStack
   │
   ▼
8. Node registers in Kubernetes
   │
   ▼
9. Pod scheduled on new node
```

---

## Directory Structure

```
karpenter-provider-cloudstack/
├── cmd/
│   └── controller/
│       └── main.go                      # Controller entrypoint
├── pkg/
│   ├── apis/                            # CRDs and API definitions
│   │   └── v1/
│   │       ├── cloudstacknodeclass.go   # CloudStackNodeClass CRD
│   │       ├── doc.go                   # API group registration
│   │       ├── labels.go                # Label definitions
│   │       └── zz_generated.deepcopy.go # Generated code
│   ├── cloudprovider/                   # Main CloudProvider implementation
│   │   └── cloudprovider.go
│   ├── cloudstack/                      # CloudStack SDK wrapper
│   │   └── sdk.go
│   ├── controllers/                     # Kubernetes controllers
│   │   ├── nodeclass/
│   │   │   └── controller.go            # NodeClass controller
│   │   └── controllers.go
│   ├── operator/                        # Operator setup
│   │   ├── operator.go
│   │   └── options/
│   │       └── options.go
│   ├── providers/                       # Cloud provider implementations
│   │   ├── instance/
│   │   │   └── instance.go              # VM lifecycle management
│   │   ├── instancetype/
│   │   │   └── instancetype.go          # Service Offering discovery
│   │   ├── network/
│   │   │   └── network.go               # Network management
│   │   ├── template/
│   │   │   └── template.go              # Template discovery
│   │   └── zone/
│   │       └── zone.go                  # Zone management
│   └── fake/                            # Mock implementations for testing
│       └── cloudstackapi.go
├── charts/                              # Helm chart
│   └── karpenter-cloudstack/
│       ├── Chart.yaml
│       ├── values.yaml
│       ├── crds/                        # CRD definitions
│       └── templates/                   # Kubernetes manifests
├── docs/                                # Documentation
│   ├── ARCHITECTURE.md                  # This file
│   ├── RELEASE.md                       # Release process
│   └── VERSIONING.md                    # Versioning guide
├── .github/
│   └── workflows/                       # CI/CD workflows
│       ├── ci.yaml                      # Continuous integration
│       ├── release.yaml                 # Release automation
│       └── README.md                    # Workflows documentation
├── Dockerfile                           # Multi-arch container image
├── Makefile                             # Build automation
├── go.mod                               # Go dependencies
├── README.md                            # Main documentation
└── INSTALLATION.md                      # Installation guide
```

---

## Integration with Karpenter Core

### Dependencies
```
sigs.k8s.io/karpenter               # Karpenter Core APIs
github.com/awslabs/operatorpkg      # Operator utilities
github.com/apache/cloudstack-go/v2  # CloudStack SDK
```

### Extension Points
Karpenter Core provides interfaces that this provider implements:

1. **cloudprovider.CloudProvider**: Main interface
2. **v1.NodeClaim**: Cloud-agnostic node abstraction
3. **v1.NodePool**: Node pool definition
4. **cloudprovider.InstanceType**: Cloud-agnostic instance type

---

## Security

### Secrets Management
CloudStack credentials are managed through Kubernetes Secrets:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: cloudstack-credentials
  namespace: karpenter
type: Opaque
stringData:
  api-url: https://cloudstack.example.com/client/api
  api-key: your-api-key
  secret-key: your-secret-key
```

### RBAC
The controller requires minimal permissions:
- Read/write CloudStackNodeClass
- Read NodeClaims and NodePools
- Read Secrets (only cloudstack-credentials)

### Network Security
- All communication with CloudStack API uses HTTPS
- Supports TLS certificate verification
- Optional: Proxy support for corporate environments

---

## Observability

### Metrics (Future)
- Number of VMs created/deleted
- VM provisioning time
- CloudStack API errors
- Cache hits/misses

### Logging
- Configurable log level (debug, info, warn, error)
- Structured JSON logs
- Request tracing context

### Health Checks
- Liveness probe: Controller is running
- Readiness probe: Can communicate with CloudStack API

---

## Future Roadmap

### Planned Features
1. **Affinity/Anti-affinity**: Support for VM anti-affinity
2. **Spot instances**: Support for CloudStack preemptible instances
3. **GPU support**: Provision VMs with GPUs
4. **Custom networking**: Support for multiple NICs
5. **Storage options**: Additional volumes
6. **Metrics exporter**: Prometheus metrics
7. **Drift detection**: Detect manual changes in VMs
8. **Cost optimization**: Cost-saving strategies

### Technical Improvements
1. **Integration tests**: Complete suite with CloudStack simulator
2. **E2E tests**: End-to-end tests on real cluster
3. **Performance profiling**: Performance optimization
4. **Cache layer**: Improved caching with configurable TTL
5. **Webhooks**: Validating/Mutating webhooks for CRDs
