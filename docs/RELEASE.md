# Release Process

Este documento describe el proceso de publicación de nuevas versiones del Karpenter Provider para CloudStack.

## 📦 Artefactos Publicados

Cada release genera los siguientes artefactos:

### 1. **Container Image**
```
ghcr.io/mperea/cloudstack/karpenter/controller:v{VERSION}
```
- Publicado en GitHub Container Registry (GHCR)
- Arquitecturas: `amd64` (arm64 en el futuro)
- Tags generados:
  - `v{VERSION}` - Versión específica (ej: `v0.1.0`)
  - `latest` - Última versión estable (solo para releases, no pre-releases)
  - `main` - Última versión de la rama main (desarrollo continuo)

### 2. **Helm Chart**
```
ghcr.io/mperea/cloudstack/karpenter/karpenter:v{VERSION}
```
- Publicado en GHCR como OCI artifact
- También disponible en GitHub Pages: `https://mperea.github.io/karpenter-provider-cloudstack`

### 3. **CRDs (Custom Resource Definitions)**
- Incluidos en el Helm Chart
- También disponibles como archivos YAML en GitHub Releases

### 4. **GitHub Release**
- Release notes generados automáticamente
- Assets adjuntos:
  - Helm Chart (`.tgz`)
  - CRDs (`.yaml`)
  - Documentación de instalación

---

## 🔄 Tipos de Releases

### **Desarrollo (main branch)**
```bash
git push origin main
```
**Artefactos generados:**
- ✅ Container Image: `:main`
- ❌ Helm Chart: No publicado
- ❌ GitHub Release: No

**Uso:** Testing continuo, desarrollo activo

---

### **Pre-Release (alpha/beta/rc)**
```bash
git tag v0.1.0-alpha.1
git push origin v0.1.0-alpha.1
```

**Artefactos generados:**
- ✅ Container Image: `:v0.1.0-alpha.1`
- ✅ Helm Chart: `0.1.0-alpha.1`
- ✅ GitHub Pre-Release (marcado como pre-release)
- ✅ CRDs en assets

**Uso:** Testing en staging, validación antes de release estable

**Instalación:**
```bash
# Helm Chart desde OCI
helm install karpenter oci://ghcr.io/mperea/cloudstack/karpenter/karpenter \
  --version 0.1.0-alpha.1

# Helm Chart desde GitHub Pages
helm repo add cloudstack-karpenter https://mperea.github.io/karpenter-provider-cloudstack
helm install karpenter cloudstack-karpenter/karpenter --version 0.1.0-alpha.1
```

---

### **Release Estable (vX.Y.Z)**
```bash
git tag v0.1.0
git push origin v0.1.0
```

**Artefactos generados:**
- ✅ Container Image: `:v0.1.0` + `:latest`
- ✅ Helm Chart: `0.1.0`
- ✅ GitHub Release (estable, release notes completos)
- ✅ CRDs en assets
- ✅ Actualización de documentación

**Uso:** Producción

**Instalación:**
```bash
# Helm Chart desde OCI
helm install karpenter oci://ghcr.io/mperea/cloudstack/karpenter/karpenter \
  --version 0.1.0

# Helm Chart desde GitHub Pages (recomendado)
helm repo add cloudstack-karpenter https://mperea.github.io/karpenter-provider-cloudstack
helm install karpenter cloudstack-karpenter/karpenter --version 0.1.0
```

---

## 📋 Versionado Semántico

Seguimos **[Semantic Versioning 2.0.0](https://semver.org/)**:

```
v{MAJOR}.{MINOR}.{PATCH}[-{PRERELEASE}]
```

### **MAJOR (v1.0.0)**
Breaking changes - Cambios incompatibles con versiones anteriores:
- Cambios en la API del CRD CloudStackNodeClass
- Eliminación de campos o funcionalidades
- Cambios en el comportamiento por defecto que rompen compatibilidad

### **MINOR (v0.1.0)**
Nuevas funcionalidades compatibles:
- Nuevos campos en CloudStackNodeClass
- Nuevas características opcionales
- Mejoras de rendimiento sin breaking changes

### **PATCH (v0.1.1)**
Bugfixes y mejoras menores:
- Corrección de errores
- Actualizaciones de seguridad
- Mejoras de documentación
- Actualizaciones de dependencias (sin breaking changes)

### **PRE-RELEASE**
Versiones no estables:
- `-alpha.N` - Versión alpha, puede tener bugs, API puede cambiar
- `-beta.N` - Versión beta, API más estable, testing en staging
- `-rc.N` - Release candidate, candidato a release estable

**Ejemplos:**
```
v0.1.0         → Primera versión estable (alpha del proyecto)
v0.1.1         → Bugfix
v0.2.0         → Nueva funcionalidad
v1.0.0         → Primera versión production-ready
v0.1.0-alpha.1 → Pre-release alpha
v0.1.0-beta.1  → Pre-release beta
v0.1.0-rc.1    → Release candidate
```

---

## 🚀 Proceso de Release Manual

### **1. Preparación**

```bash
# Actualizar rama main
git checkout main
git pull origin main

# Crear rama de release (opcional, para cambios de última hora)
git checkout -b release/v0.1.0

# Actualizar CHANGELOG (si existe)
# Actualizar versiones en Chart.yaml si es necesario
```

### **2. Verificación Pre-Release**

```bash
# Ejecutar todos los tests
make test

# Verificar linting
make lint

# Build local
make build

# Verificar que todo compila
make docker-build
```

### **3. Crear Tag**

```bash
# Para release estable
git tag -a v0.1.0 -m "Release v0.1.0"

# Para pre-release
git tag -a v0.1.0-alpha.1 -m "Pre-release v0.1.0-alpha.1"

# Verificar el tag
git tag -l -n9 v0.1.0
```

### **4. Push del Tag**

```bash
# Push el tag (esto dispara el workflow de release)
git push origin v0.1.0
```

### **5. Monitorear el Pipeline**

```bash
# El workflow de GitHub Actions se ejecuta automáticamente
# Verificar en: https://github.com/mperea/karpenter-provider-cloudstack/actions
```

### **6. Verificar Artefactos**

```bash
# Verificar imagen Docker
docker pull ghcr.io/mperea/cloudstack/karpenter/controller:v0.1.0

# Verificar Helm Chart (OCI)
helm pull oci://ghcr.io/mperea/cloudstack/karpenter/karpenter --version 0.1.0

# Verificar GitHub Release
# https://github.com/mperea/karpenter-provider-cloudstack/releases
```

### **7. Actualizar Documentación**

```bash
# Actualizar README con la nueva versión
# Actualizar INSTALLATION.md si hay cambios
# Anunciar en canales relevantes (Slack, Twitter, etc.)
```

---

## 🛠️ Rollback de Release

Si necesitas revertir un release:

```bash
# Eliminar el tag localmente
git tag -d v0.1.0

# Eliminar el tag del remoto
git push origin :refs/tags/v0.1.0

# Eliminar la GitHub Release (desde la UI o con gh CLI)
gh release delete v0.1.0

# Eliminar las imágenes de GHCR (desde la UI)
# https://github.com/users/mperea/packages/container/cloudstack%2Fkarpenter%2Fcontroller
```

---

## 📝 Checklist de Release

### **Pre-Release**
- [ ] Todos los tests pasan (`make test`)
- [ ] Linting limpio (`make lint`)
- [ ] Documentación actualizada
- [ ] CHANGELOG actualizado (si aplica)
- [ ] Versión actualizada en Chart.yaml (si es manual)
- [ ] Branch main actualizado

### **Release**
- [ ] Tag creado con nombre correcto (vX.Y.Z o vX.Y.Z-alpha.N)
- [ ] Tag pusheado a GitHub
- [ ] Pipeline de CI/CD completado exitosamente
- [ ] Artefactos verificados (imagen, chart, GitHub release)

### **Post-Release**
- [ ] Instalación probada desde Helm
- [ ] Documentación verificada
- [ ] Release anunciado (si es relevante)
- [ ] Issues/PRs actualizados con la nueva versión

---

## 🔗 Referencias

- [Semantic Versioning 2.0.0](https://semver.org/)
- [Karpenter Documentation](https://karpenter.sh)
- [Helm Best Practices](https://helm.sh/docs/chart_best_practices/)
- [GitHub Container Registry](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry)

