# Project Versioning

This project follows **[Semantic Versioning 2.0.0](https://semver.org/spec/v2.0.0.html)** (Semver 2.0).

---

## Version Format

```
vMAJOR.MINOR.PATCH[-PRERELEASE][+BUILD]
```

### Required Components

#### MAJOR (vX.0.0)
Increment when making **incompatible API changes**:
- Breaking changes in CloudStackNodeClass CRD
- Removal of fields or functionality
- Changes requiring manual migration
- Changes in behavior that break compatibility

**Examples:**
- Remove `spec.userData` field from CloudStackNodeClass
- Change field type from `string` to `object`
- Change default behavior in incompatible ways

#### MINOR (v0.X.0)
Increment when adding **new backwards-compatible functionality**:
- New optional fields in CloudStackNodeClass
- New features that don't break compatibility
- Significant performance improvements
- New providers or components

**Examples:**
- Add optional `spec.diskOffering` field
- Add support for multiple NICs
- Implement Service Offerings cache

#### PATCH (v0.0.X)
Increment for **bug fixes** and minor backwards-compatible changes:
- Error corrections
- Security updates
- Documentation improvements
- Dependency updates (without breaking changes)
- Code optimizations without functional changes

**Examples:**
- Fix memory leak in cache
- Update dependency with security fix
- Fix tag validation

---

### Optional Components

#### PRERELEASE (v0.1.0-alpha.1)
Identifiers for unstable versions:

- **`alpha.N`**: Early development, may have bugs, unstable API
- **`beta.N`**: Feature complete, testing phase, more stable API
- **`rc.N`**: Release candidate, production-ready testing

**Rules:**
- Identifiers are alphanumeric with hyphens: `[0-9A-Za-z-]`
- Multiple identifiers separated by dots
- No leading zeros in numeric identifiers

**Valid examples:**
```
v0.1.0-alpha.1
v0.1.0-alpha.2
v0.1.0-beta.1
v0.1.0-rc.1
v1.0.0-alpha
v1.0.0-alpha.20130313
v1.0.0-0.3.7
```

**Precedence:**
```
v1.0.0-alpha < v1.0.0-alpha.1 < v1.0.0-beta < v1.0.0-beta.2 < v1.0.0-rc.1 < v1.0.0
```

#### BUILD (v0.1.0+20130313144700)
Build metadata (does NOT affect version precedence):

- Commit SHA information
- Timestamps
- Build identifiers

**Rules:**
- Added with `+`
- Only alphanumeric with hyphens and dots: `[0-9A-Za-z-.]+`
- **Not used for precedence**: `v1.0.0+001 = v1.0.0+002`

**Valid examples:**
```
v1.0.0+20130313144700
v1.0.0+exp.sha.5114f85
v1.0.0-beta.1+exp.sha.5114f85
```

---

## Complete Examples

### Stable Versions
```bash
v0.1.0          # First stable version (minor)
v0.1.1          # Bugfix
v0.2.0          # New functionality
v1.0.0          # First production-ready version
v1.0.1          # Bugfix in v1
v1.1.0          # New functionality in v1
v2.0.0          # Breaking change
```

### Pre-releases
```bash
v0.1.0-alpha.1  # Alpha 1 of v0.1.0
v0.1.0-alpha.2  # Alpha 2 of v0.1.0
v0.1.0-beta.1   # Beta 1 of v0.1.0
v0.1.0-beta.2   # Beta 2 of v0.1.0
v0.1.0-rc.1     # Release candidate 1
v0.1.0-rc.2     # Release candidate 2
v0.1.0          # Final release
```

### With Build Metadata
```bash
v1.0.0+20130313144700
v1.0.0-beta.1+exp.sha.5114f85
v1.2.3-rc.1+build.123
```

### Invalid Versions
```bash
1.0.0             # Missing 'v' prefix
v1.0              # Missing PATCH component
v1.0.0-Alpha.1    # Uppercase 'A' in prerelease
v1.0.0.0          # Too many components
v01.0.0           # Leading zero in MAJOR
v1.01.0           # Leading zero in MINOR
v1.0.01           # Leading zero in PATCH
```

---

## Precedence Rules (Ordering)

Semver 2.0.0 defines strict precedence ordering:

```
v1.0.0-alpha < v1.0.0-alpha.1 < v1.0.0-alpha.beta < v1.0.0-beta <
v1.0.0-beta.2 < v1.0.0-beta.11 < v1.0.0-rc.1 < v1.0.0
```

**Rules:**
1. MAJOR, MINOR, PATCH compared numerically
2. Pre-release has lower precedence than stable version
3. Pre-release identifiers compared left-to-right:
   - Numeric compared as numbers: `1 < 2 < 10`
   - Alphanumeric compared lexicographically: `"alpha" < "beta"`
   - Numeric < Alphanumeric: `1 < alpha`
4. Build metadata does NOT affect precedence

**Example:**
```
v1.0.0-alpha.1 < v1.0.0-alpha.2 < v1.0.0-beta.1 < v1.0.0
```

---

## Release Process

### 1. Alpha Version (Initial Testing)
```bash
git tag -a v0.1.0-alpha.1 -m "Alpha 1: Initial testing"
git push origin v0.1.0-alpha.1
```
**Usage:** Internal testing, active development, API may change

### 2. Beta Version (Feature Complete)
```bash
git tag -a v0.1.0-beta.1 -m "Beta 1: Feature complete"
git push origin v0.1.0-beta.1
```
**Usage:** Staging testing, API stabilized

### 3. Release Candidate (Production Candidate)
```bash
git tag -a v0.1.0-rc.1 -m "Release Candidate 1"
git push origin v0.1.0-rc.1
```
**Usage:** Final testing before production

### 4. Stable Release (Production)
```bash
git tag -a v0.1.0 -m "Release v0.1.0"
git push origin v0.1.0
```
**Usage:** Production, `latest` tag updated

---

## Decision Guide

### When to increment each component?

```
┌─────────────────────────────────────────────────────────┐
│ Does it break compatibility with previous version?     │
│ (Breaking changes)                                      │
└────────────────┬────────────────────────────────────────┘
                 │
         ┌───────┴───────┐
         │ YES           │ NO
         │               │
         ▼               ▼
    MAJOR++      ┌──────────────────────────────────────┐
                 │ Does it add new functionality?       │
                 │ (New features)                       │
                 └────────────┬─────────────────────────┘
                              │
                      ┌───────┴───────┐
                      │ YES           │ NO
                      │               │
                      ▼               ▼
                  MINOR++        PATCH++
```

### Change Examples

| Change | Type | New Version |
|--------|------|---------------|
| Add optional field in CRD | MINOR | v0.1.0 → v0.2.0 |
| Remove CRD field | MAJOR | v0.2.0 → v1.0.0 |
| Fix bug | PATCH | v0.2.0 → v0.2.1 |
| Update docs | PATCH | v0.2.1 → v0.2.2 |
| Add new feature | MINOR | v0.2.2 → v0.3.0 |
| Incompatible API change | MAJOR | v0.3.0 → v1.0.0 |

---

## Version Validation

The CI workflow automatically validates that tags follow Semver 2.0.0:

**Validation regex:**
```regex
^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-((0|[1-9][0-9]*|[0-9]*[a-zA-Z-][0-9a-zA-Z-]*)(\.(0|[1-9][0-9]*|[0-9]*[a-zA-Z-][0-9a-zA-Z-]*))*))?(\+([0-9a-zA-Z-]+(\.[0-9a-zA-Z-]+)*))?$
```

If the tag is invalid, the build will fail with a descriptive error.

---

## References

- **Semantic Versioning 2.0.0**: https://semver.org/spec/v2.0.0.html
- **Semver Calculator**: https://semver.npmjs.com/
- **Regex101 (test regex)**: https://regex101.com/

---

## Recommendations

### Pre-1.0.0 (Initial Development)
- MINOR may have breaking changes
- Use v0.x.x until API is stable
- First stable version: v1.0.0

### Post-1.0.0 (Production)
- MAJOR only for breaking changes
- MINOR for compatible new features
- PATCH for bugfixes

### Pre-releases
- Always test in staging before stable release
- Use alpha → beta → rc → stable
- Don't skip directly to stable from alpha
