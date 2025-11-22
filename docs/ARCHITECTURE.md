# Arquitectura del Proyecto

## 📐 Visión General

Karpenter Provider for CloudStack sigue la arquitectura extensible de Karpenter, implementando los interfaces específicos para CloudStack.

```
┌─────────────────────────────────────────────────────────────┐
│                    Karpenter Core                            │
│  (Scheduling, Provisioning, Deprovisioning logic)           │
└───────────────────────┬─────────────────────────────────────┘
                        │ Cloud Provider Interface
┌───────────────────────▼─────────────────────────────────────┐
│              Karpenter CloudStack Provider                   │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │  Instance    │  │ InstanceType │  │   Network    │      │
│  │  Provider    │  │  Provider    │  │   Provider   │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │  Template    │  │    Zone      │  │ NodeClass    │      │
│  │  Provider    │  │  Provider    │  │ Controller   │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└───────────────────────┬─────────────────────────────────────┘
                        │ CloudStack SDK
┌───────────────────────▼─────────────────────────────────────┐
│                  CloudStack API                              │
│  (deployVirtualMachine, listServiceOfferings, etc.)         │
└─────────────────────────────────────────────────────────────┘
```

---

## 🏗️ Componentes Principales

### **1. Custom Resource Definitions (CRDs)**

#### **CloudStackNodeClass**
Define la configuración específica de CloudStack para los nodos:

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

**Campos principales:**
- `zone`: Zona de CloudStack donde se crearán las VMs
- `networkSelectorTerms`: Selección de red(es)
- `serviceOfferingSelectorTerms`: Selección de Service Offerings
- `templateSelectorTerms`: Selección de templates/imágenes
- `userData`: Script de inicialización
- `tags`: Tags a aplicar a las VMs

### **2. Providers**

#### **Instance Provider** (`pkg/providers/instance/`)
Gestiona el ciclo de vida de las VMs en CloudStack:

```go
type Provider interface {
    Create(context.Context, *v1.NodeClaim) (*v1.Instance, error)
    Get(context.Context, string) (*v1.Instance, error)
    List(context.Context) ([]*v1.Instance, error)
    Delete(context.Context, string) error
}
```

**Responsabilidades:**
- Crear VMs usando `deployVirtualMachine`
- Consultar estado de VMs
- Eliminar VMs usando `destroyVirtualMachine`
- Mapear VMs de CloudStack a Instances de Karpenter

#### **InstanceType Provider** (`pkg/providers/instancetype/`)
Descubre y cachea Service Offerings de CloudStack:

```go
type Provider interface {
    List(context.Context, *v1.NodeClaim) ([]*cloudprovider.InstanceType, error)
    Get(context.Context, string) (*cloudprovider.InstanceType, error)
}
```

**Responsabilidades:**
- Listar Service Offerings usando `listServiceOfferings`
- Convertir Service Offerings a InstanceTypes de Karpenter
- Cachear resultados para mejorar rendimiento
- Filtrar por tags y requisitos

#### **Network Provider** (`pkg/providers/network/`)
Gestiona la selección de redes:

```go
type Provider interface {
    GetByName(context.Context, string, string) (*cloudstack.Network, error)
    GetByID(context.Context, string) (*cloudstack.Network, error)
    GetByTags(context.Context, string, map[string]string) ([]*cloudstack.Network, error)
}
```

**Responsabilidades:**
- Buscar redes por nombre, ID o tags
- Validar disponibilidad de redes
- Cachear información de redes

#### **Template Provider** (`pkg/providers/template/`)
Gestiona la selección de templates/imágenes:

```go
type Provider interface {
    GetByName(context.Context, string, string) (*cloudstack.Template, error)
    GetByID(context.Context, string) (*cloudstack.Template, error)
    GetByTags(context.Context, string, map[string]string) ([]*cloudstack.Template, error)
}
```

**Responsabilidades:**
- Buscar templates por nombre, ID o tags
- Validar templates disponibles
- Cachear información de templates

#### **Zone Provider** (`pkg/providers/zone/`)
Gestiona información de zonas de CloudStack:

```go
type Provider interface {
    List(context.Context) ([]*cloudstack.Zone, error)
    GetByName(context.Context, string) (*cloudstack.Zone, error)
}
```

**Responsabilidades:**
- Listar zonas disponibles
- Validar zonas
- Cachear información de zonas

### **3. Controllers**

#### **NodeClass Controller** (`pkg/controllers/nodeclass/`)
Reconcilia CloudStackNodeClass recursos:

**Responsabilidades:**
- Validar configuración de CloudStackNodeClass
- Verificar que zone, networks, templates existan
- Actualizar status conditions (Ready, NetworkReady, etc.)
- Detectar cambios de configuración (drift)

### **4. Cloud Provider** (`pkg/cloudprovider/`)
Implementa la interfaz principal de Karpenter:

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

**Responsabilidades:**
- Orquestar los diferentes providers
- Implementar la lógica de creación/eliminación de nodos
- Gestionar el mapeo entre Karpenter y CloudStack
- Generar ProviderID único para cada nodo

---

## 🔄 Flujo de Provisión de Nodo

```
1. Pod sin schedulear
   │
   ▼
2. Karpenter Core detecta necesidad
   │
   ▼
3. Calcula requisitos (CPU, RAM, labels, taints)
   │
   ▼
4. CloudProvider.GetInstanceTypes()
   │
   ├─► InstanceType Provider lista Service Offerings
   │   └─► Filtra por requisitos
   │
   ▼
5. Selecciona InstanceType óptimo
   │
   ▼
6. CloudProvider.Create(NodeClaim)
   │
   ├─► Template Provider busca imagen
   ├─► Network Provider busca red
   └─► Instance Provider crea VM
       │
       └─► cloudstack.deployVirtualMachine()
   │
   ▼
7. VM creada en CloudStack
   │
   ▼
8. Node se registra en Kubernetes
   │
   ▼
9. Pod se schedule en el nuevo nodo
```

---

## 🗂️ Estructura de Directorios

```
karpenter-provider-cloudstack/
├── cmd/
│   └── controller/
│       └── main.go                      # Entrypoint del controller
├── pkg/
│   ├── apis/                            # CRDs y API definitions
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
│   └── RELEASE.md                       # Release process
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

## 🔌 Integración con Karpenter Core

### **Dependencias**
```
sigs.k8s.io/karpenter               # Karpenter Core APIs
github.com/awslabs/operatorpkg      # Operator utilities
github.com/apache/cloudstack-go/v2  # CloudStack SDK
```

### **Puntos de Extensión**
Karpenter Core proporciona interfaces que este provider implementa:

1. **cloudprovider.CloudProvider**: Interface principal
2. **v1.NodeClaim**: Abstracción de nodo cloud-agnostic
3. **v1.NodePool**: Definición de pool de nodos
4. **cloudprovider.InstanceType**: Tipo de instancia cloud-agnostic

---

## 🛡️ Seguridad

### **Secrets Management**
Las credenciales de CloudStack se gestionan mediante Kubernetes Secrets:

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

### **RBAC**
El controller requiere permisos mínimos:
- Leer/escribir CloudStackNodeClass
- Leer NodeClaims y NodePools
- Leer Secrets (solo cloudstack-credentials)

### **Network Security**
- Todas las comunicaciones con CloudStack API usan HTTPS
- Soporta verificación de certificados TLS
- Opcional: Proxy support para entornos corporativos

---

## 📊 Observabilidad

### **Métricas** (Futuro)
- Número de VMs creadas/eliminadas
- Tiempo de provisión de VMs
- Errores de API de CloudStack
- Cache hits/misses

### **Logging**
- Nivel de log configurable (debug, info, warn, error)
- Logs estructurados en JSON
- Contexto de request tracing

### **Health Checks**
- Liveness probe: Controller está corriendo
- Readiness probe: Puede comunicar con CloudStack API

---

## 🔮 Roadmap Futuro

### **Features Planeadas**
1. **Affinity/Anti-affinity**: Soporte para anti-affinity entre VMs
2. **Spot instances**: Soporte para CloudStack preemptible instances
3. **GPU support**: Provisión de VMs con GPUs
4. **Custom networking**: Soporte para múltiples NICs
5. **Storage options**: Volúmenes adicionales
6. **Metrics exporter**: Prometheus metrics
7. **Drift detection**: Detectar cambios manuales en VMs
8. **Cost optimization**: Estrategias de ahorro de costos

### **Mejoras Técnicas**
1. **Tests de integración**: Suite completa con CloudStack simulator
2. **E2E tests**: Tests end-to-end en cluster real
3. **Performance profiling**: Optimización de rendimiento
4. **Cache layer**: Mejora de caching con TTL configurable
5. **Webhooks**: Validating/Mutating webhooks para CRDs

