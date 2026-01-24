# GitHub Actions Workflows

## Workflow: Build and Release

### Descripción

Este workflow se ejecuta automáticamente cuando se hace push a la rama `master` y:

1. **Construye el binario Go** del metrics server
2. **Construye la extensión UI** (React/TypeScript)
3. **Crea un release de GitHub** con los artefactos

### Trigger

- **Rama**: `master`
- **Evento**: `push`
- **Ignora cambios en**: archivos `.md`, `.gitignore`, `LICENSE`

### Proceso

1. **Setup**:
   - Checkout del código
   - Instalación de Go 1.21
   - Instalación de Node.js 18
   - Instalación de Yarn

2. **Build**:
   - Construcción del binario Go (`argocd-metrics-server`)
   - Construcción de la extensión UI (`extension.tar.gz`)
   - Generación de checksums SHA256

3. **Release**:
   - Si hay un tag en el commit → Release oficial
   - Si no hay tag → Pre-release con versión basada en fecha y commit

### Versiones

- **Con tag**: Usa el tag como versión (ej: `v1.0.3`)
- **Sin tag**: Crea versión `vYYYYMMDD-<short-commit>` (ej: `v20241215-a1b2c3d`)

### Artifacts Generados

1. **`argocd-metrics-server`**: Binario compilado para Linux
2. **`extension.tar.gz`**: Extensión UI comprimida para ArgoCD
3. **`extension_checksums.txt`**: Checksums SHA256 para verificación

### Uso de los Artifacts

#### Metrics Server

```bash
# Descargar desde el release
wget https://github.com/<repo>/releases/download/<version>/argocd-metrics-server
chmod +x argocd-metrics-server
```

#### UI Extension

```yaml
# En el deployment de ArgoCD
initContainers:
  - name: extension-metrics
    image: quay.io/argoprojlabs/argocd-extension-installer:v0.0.1
    env:
    - name: EXTENSION_URL
      value: https://github.com/<repo>/releases/download/<version>/extension.tar.gz
    - name: EXTENSION_CHECKSUM_URL
      value: https://github.com/<repo>/releases/download/<version>/extension_checksums.txt
```

### Permisos Requeridos

El workflow necesita el permiso `GITHUB_TOKEN` para crear releases. Este token se proporciona automáticamente por GitHub Actions.

### Ver el Workflow

El workflow está en: `.github/workflows/release.yml`

### Ejecución Manual

Si necesitas ejecutar el workflow manualmente, puedes usar:

```bash
# Crear un tag para trigger un release oficial
git tag v1.0.4
git push origin v1.0.4

# O hacer push a master para un pre-release
git push origin master
```
