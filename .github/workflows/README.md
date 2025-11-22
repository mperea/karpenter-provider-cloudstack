# GitHub Actions Workflows

Este directorio contiene los workflows de CI/CD del proyecto.

## Workflows Disponibles

### 1. **CI (ci.yaml)**

Workflow de integración continua que se ejecuta en cada push y pull request.

**Triggers:**
- Push a `main`
- Pull requests a `main`

**Jobs:**
- **test**: Tests unitarios y linting
- **build**: Compilación del binario
- **docker**: Build de imagen Docker (solo en push)

### 2. **Release (release.yaml)**

Workflow de publicación automática que se ejecuta al crear tags.

**Triggers:**
- Push de tags: `v*` (ej: `v0.1.0`, `v0.1.0-alpha.1`)

**Jobs:**
- **release**:
  - Build multi-arch container image (amd64, arm64)
  - Push a GitHub Container Registry
  - Package Helm chart
  - Push Helm chart a OCI registry
  - Crear GitHub Release con assets
  
- **update-helm-repo**:
  - Actualizar repositorio Helm en GitHub Pages

**Artefactos generados:**
- 🐳 Container Image: `ghcr.io/mperea/cloudstack/karpenter/controller:v{VERSION}`
- 📦 Helm Chart (OCI): `oci://ghcr.io/mperea/cloudstack/karpenter/karpenter`
- 📦 Helm Chart (GitHub Pages): `https://mperea.github.io/karpenter-provider-cloudstack`
- 📝 GitHub Release con CRDs

## Crear un Release

```bash
# 1. Asegúrate de estar en main actualizado
git checkout main
git pull origin main

# 2. Crea y push el tag
git tag -a v0.1.0 -m "Release v0.1.0"
git push origin v0.1.0

# 3. El workflow se ejecuta automáticamente
# Monitorea en: https://github.com/mperea/karpenter-provider-cloudstack/actions
```

## Variables de Entorno

### CI Workflow
- `go-version`: Versión de Go (1.25.4)

### Release Workflow
- `REGISTRY`: GitHub Container Registry (`ghcr.io`)
- `IMAGE_NAME`: Nombre de la imagen (`cloudstack/karpenter/controller`)
- `CHART_NAME`: Nombre del chart (`cloudstack/karpenter/karpenter`)

## Permisos Requeridos

Los workflows requieren los siguientes permisos en el token `GITHUB_TOKEN`:

- `contents: write` - Para crear GitHub Releases
- `packages: write` - Para publicar en GHCR

Estos permisos están configurados automáticamente en el workflow.

## Troubleshooting

### Error: "Resource not accessible by integration"
**Causa:** Falta permiso `packages: write`  
**Solución:** Verificar que el workflow tiene `permissions: packages: write`

### Error: "Image not found" al instalar
**Causa:** La imagen no se publicó correctamente  
**Solución:** Verificar que el workflow `release` completó exitosamente

### Error: "gh-pages branch not found"
**Causa:** La rama `gh-pages` no existe  
**Solución:** Crear manualmente la rama:
```bash
git checkout --orphan gh-pages
git rm -rf .
echo "# Helm Charts" > README.md
git add README.md
git commit -m "Initialize gh-pages"
git push origin gh-pages
```

Luego configurar en GitHub Settings → Pages → Source: `gh-pages` branch.
