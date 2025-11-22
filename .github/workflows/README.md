# GitHub Actions Workflows

This directory contains the CI/CD workflows for the project.

## Available Workflows

### 1. CI (ci.yaml)

Continuous integration workflow that runs on every push and pull request.

**Triggers:**
- Push to `main`
- Pull requests to `main`

**Jobs:**
- **test**: Unit tests and linting
- **build**: Binary compilation
- **docker**: Docker image build (push only)

### 2. Release (release.yaml)

Automatic publication workflow that runs when creating tags.

**Triggers:**
- Push tags: `v*` (e.g., `v0.1.0`, `v0.1.0-alpha.1`)

**Jobs:**
- **release**:
  - Multi-arch container image build (amd64, arm64)
  - Push to GitHub Container Registry
  - Package Helm chart
  - Push Helm chart to OCI registry
  - Create GitHub Release with assets

- **update-helm-repo**:
  - Update Helm repository on GitHub Pages

**Generated artifacts:**
- Container Image: `ghcr.io/mperea/cloudstack/karpenter/controller:v{VERSION}`
- Helm Chart (OCI): `oci://ghcr.io/mperea/cloudstack/karpenter/karpenter`
- Helm Chart (GitHub Pages): `https://mperea.github.io/karpenter-provider-cloudstack`
- GitHub Release with CRDs

## Creating a Release

```bash
# 1. Ensure you're on updated main
git checkout main
git pull origin main

# 2. Create and push the tag
git tag -a v0.1.0 -m "Release v0.1.0"
git push origin v0.1.0

# 3. The workflow runs automatically
# Monitor at: https://github.com/mperea/karpenter-provider-cloudstack/actions
```

## Environment Variables

### CI Workflow
- `go-version`: Go version (1.25.4)

### Release Workflow
- `REGISTRY`: GitHub Container Registry (`ghcr.io`)
- `IMAGE_NAME`: Image name (`cloudstack/karpenter/controller`)
- `CHART_NAME`: Chart name (`cloudstack/karpenter/karpenter`)

## Required Permissions

Workflows require the following permissions on the `GITHUB_TOKEN`:

- `contents: write` - To create GitHub Releases
- `packages: write` - To publish to GHCR

These permissions are automatically configured in the workflow.

## Troubleshooting

### Error: "Resource not accessible by integration"
**Cause:** Missing `packages: write` permission
**Solution:** Verify that the workflow has `permissions: packages: write`

### Error: "Image not found" when installing
**Cause:** Image was not published correctly
**Solution:** Verify that the `release` workflow completed successfully

### Error: "gh-pages branch not found"
**Cause:** The `gh-pages` branch doesn't exist
**Solution:** Create the branch manually:
```bash
git checkout --orphan gh-pages
git rm -rf .
echo "# Helm Charts" > README.md
git add README.md
git commit -m "Initialize gh-pages"
git push origin gh-pages
```

Then configure in GitHub Settings → Pages → Source: `gh-pages` branch.
