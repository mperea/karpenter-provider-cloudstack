# Versionado del Proyecto

Este proyecto sigue **[Semantic Versioning 2.0.0](https://semver.org/spec/v2.0.0.html)** (Semver 2.0).

---

## 📐 Formato de Versión

```
vMAJOR.MINOR.PATCH[-PRERELEASE][+BUILD]
```

### **Componentes Obligatorios**

#### **MAJOR** (v**X**.0.0)
Incrementar cuando hay **cambios incompatibles** en la API:
- Breaking changes en CloudStackNodeClass CRD
- Eliminación de campos o funcionalidades
- Cambios que requieren migración manual
- Cambios en el comportamiento que rompen compatibilidad

**Ejemplos:**
- Eliminar campo `spec.userData` de CloudStackNodeClass
- Cambiar tipo de campo de `string` a `object`
- Cambiar comportamiento por defecto de forma incompatible

#### **MINOR** (v0.**X**.0)
Incrementar cuando se añaden **nuevas funcionalidades** de forma compatible:
- Nuevos campos opcionales en CloudStackNodeClass
- Nuevas características que no rompen compatibilidad
- Mejoras significativas de rendimiento
- Nuevos providers o componentes

**Ejemplos:**
- Añadir campo opcional `spec.diskOffering`
- Añadir soporte para múltiples NICs
- Implementar cache de Service Offerings

#### **PATCH** (v0.0.**X**)
Incrementar para **bugfixes** y cambios menores compatibles:
- Corrección de errores
- Actualizaciones de seguridad
- Mejoras de documentación
- Actualizaciones de dependencias (sin breaking changes)
- Optimizaciones de código sin cambio funcional

**Ejemplos:**
- Corregir memory leak en cache
- Actualizar dependencia con fix de seguridad
- Corregir validación de tags

---

### **Componentes Opcionales**

#### **PRERELEASE** (v0.1.0-**alpha.1**)
Identificadores de versiones no estables:

- **`alpha.N`**: Desarrollo temprano, puede tener bugs, API inestable
- **`beta.N`**: Feature complete, fase de testing, API más estable
- **`rc.N`**: Release candidate, candidato para release estable

**Reglas:**
- Los identificadores son alfanuméricos con guiones: `[0-9A-Za-z-]`
- Se separan con puntos para múltiples identificadores
- No hay ceros a la izquierda en identificadores numéricos

**Ejemplos válidos:**
```
v0.1.0-alpha.1
v0.1.0-alpha.2
v0.1.0-beta.1
v0.1.0-rc.1
v1.0.0-alpha
v1.0.0-alpha.20130313
v1.0.0-0.3.7
```

**Precedencia:**
```
v1.0.0-alpha < v1.0.0-alpha.1 < v1.0.0-beta < v1.0.0-beta.2 < v1.0.0-rc.1 < v1.0.0
```

#### **BUILD** (v0.1.0+**20130313144700**)
Metadatos de build (NO afecta precedencia de versión):

- Información de commit SHA
- Timestamps
- Identificadores de build

**Reglas:**
- Se añade con `+`
- Solo alfanumérico con guiones y puntos: `[0-9A-Za-z-.]+`
- **No se usa para precedencia**: `v1.0.0+001 = v1.0.0+002`

**Ejemplos válidos:**
```
v1.0.0+20130313144700
v1.0.0+exp.sha.5114f85
v1.0.0-beta.1+exp.sha.5114f85
```

---

## 📊 Ejemplos Completos

### **Versiones Estables**
```bash
v0.1.0          # Primera versión estable (minor)
v0.1.1          # Bugfix
v0.2.0          # Nueva funcionalidad
v1.0.0          # Primera versión production-ready
v1.0.1          # Bugfix en v1
v1.1.0          # Nueva funcionalidad en v1
v2.0.0          # Breaking change
```

### **Pre-releases**
```bash
v0.1.0-alpha.1  # Alpha 1 de v0.1.0
v0.1.0-alpha.2  # Alpha 2 de v0.1.0
v0.1.0-beta.1   # Beta 1 de v0.1.0
v0.1.0-beta.2   # Beta 2 de v0.1.0
v0.1.0-rc.1     # Release candidate 1
v0.1.0-rc.2     # Release candidate 2
v0.1.0          # Release final
```

### **Con Build Metadata**
```bash
v1.0.0+20130313144700
v1.0.0-beta.1+exp.sha.5114f85
v1.2.3-rc.1+build.123
```

### **Versiones Inválidas** ❌
```bash
1.0.0             # ❌ Falta prefijo 'v'
v1.0              # ❌ Falta componente PATCH
v1.0.0-Alpha.1    # ❌ 'A' mayúscula en prerelease
v1.0.0.0          # ❌ Demasiados componentes
v01.0.0           # ❌ Cero a la izquierda en MAJOR
v1.01.0           # ❌ Cero a la izquierda en MINOR
v1.0.01           # ❌ Cero a la izquierda en PATCH
```

---

## 🔄 Reglas de Precedencia (Orden)

Semver 2.0.0 define un orden estricto de precedencia:

```
v1.0.0-alpha < v1.0.0-alpha.1 < v1.0.0-alpha.beta < v1.0.0-beta <
v1.0.0-beta.2 < v1.0.0-beta.11 < v1.0.0-rc.1 < v1.0.0
```

**Reglas:**
1. MAJOR, MINOR, PATCH se comparan numéricamente
2. Pre-release tiene menor precedencia que versión estable
3. Identificadores de pre-release se comparan de izquierda a derecha:
   - Numéricos se comparan como números: `1 < 2 < 10`
   - Alfanuméricos se comparan lexicográficamente: `"alpha" < "beta"`
   - Numérico < Alfanumérico: `1 < alpha`
4. Build metadata NO afecta precedencia

**Ejemplo:**
```
v1.0.0-alpha.1 < v1.0.0-alpha.2 < v1.0.0-beta.1 < v1.0.0
```

---

## 🚀 Proceso de Release

### **1. Versión Alpha (Testing Inicial)**
```bash
git tag -a v0.1.0-alpha.1 -m "Alpha 1: Initial testing"
git push origin v0.1.0-alpha.1
```
**Uso:** Testing interno, desarrollo activo, API puede cambiar

### **2. Versión Beta (Feature Complete)**
```bash
git tag -a v0.1.0-beta.1 -m "Beta 1: Feature complete"
git push origin v0.1.0-beta.1
```
**Uso:** Testing en staging, API estabilizada

### **3. Release Candidate (Candidato a Producción)**
```bash
git tag -a v0.1.0-rc.1 -m "Release Candidate 1"
git push origin v0.1.0-rc.1
```
**Uso:** Testing final antes de producción

### **4. Release Estable (Producción)**
```bash
git tag -a v0.1.0 -m "Release v0.1.0"
git push origin v0.1.0
```
**Uso:** Producción, tag `latest` se actualiza

---

## 🔖 Guía de Decisión

### **¿Cuándo incrementar cada componente?**

```
┌─────────────────────────────────────────────────────────┐
│ ¿Rompe compatibilidad con versión anterior?             │
│ (Breaking changes)                                       │
└────────────────┬────────────────────────────────────────┘
                 │
         ┌───────┴───────┐
         │ SÍ            │ NO
         │               │
         ▼               ▼
    MAJOR++      ┌──────────────────────────────────────┐
                 │ ¿Añade nueva funcionalidad?          │
                 │ (New features)                       │
                 └────────────┬─────────────────────────┘
                              │
                      ┌───────┴───────┐
                      │ SÍ            │ NO
                      │               │
                      ▼               ▼
                  MINOR++        PATCH++
```

### **Ejemplos de Cambios**

| Cambio | Tipo | Nueva Versión |
|--------|------|---------------|
| Añadir campo opcional en CRD | MINOR | v0.1.0 → v0.2.0 |
| Eliminar campo de CRD | MAJOR | v0.2.0 → v1.0.0 |
| Corregir bug | PATCH | v0.2.0 → v0.2.1 |
| Actualizar docs | PATCH | v0.2.1 → v0.2.2 |
| Añadir nueva feature | MINOR | v0.2.2 → v0.3.0 |
| Cambiar API incompatible | MAJOR | v0.3.0 → v1.0.0 |

---

## ✅ Validación de Versiones

El workflow de CI valida automáticamente que los tags sigan Semver 2.0.0:

**Regex de validación:**
```regex
^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-((0|[1-9][0-9]*|[0-9]*[a-zA-Z-][0-9a-zA-Z-]*)(\.(0|[1-9][0-9]*|[0-9]*[a-zA-Z-][0-9a-zA-Z-]*))*))?(\+([0-9a-zA-Z-]+(\.[0-9a-zA-Z-]+)*))?$
```

Si el tag no es válido, el build fallará con un error descriptivo.

---

## 📚 Referencias

- **Semantic Versioning 2.0.0**: https://semver.org/spec/v2.0.0.html
- **Semver Calculator**: https://semver.npmjs.com/
- **Regex101 (testar regex)**: https://regex101.com/

---

## 🎯 Recomendaciones

### **Pre-1.0.0 (Desarrollo Inicial)**
- MINOR puede tener breaking changes
- Usar v0.x.x hasta que la API sea estable
- Primera versión estable: v1.0.0

### **Post-1.0.0 (Producción)**
- MAJOR solo para breaking changes
- MINOR para nuevas features compatibles
- PATCH para bugfixes

### **Pre-releases**
- Siempre testear en staging antes de release estable
- Usar alpha → beta → rc → stable
- No saltar directamente a stable desde alpha

