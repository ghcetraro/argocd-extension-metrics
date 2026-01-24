# Modificación: Conexión Directa al Metrics Server

## ⚠️ IMPORTANTE: Configuración de CORS Requerida

**Antes de usar esta modificación**, asegúrate de que el metrics server tenga CORS configurado para permitir peticiones desde `https://argocd-dev.devops.ligo.live`.

Sin CORS, el navegador bloqueará las peticiones con errores como:
```
Access to fetch at 'http://argocd-metrics-dev.devops.ligo.live/...' from origin 'https://argocd-dev.devops.ligo.live' has been blocked by CORS policy
```

**Solución**: Configurar CORS en el metrics server o agregar headers CORS en el ALB/Ingress.

## Cambios Realizados

Se modificó la extensión UI para que se conecte directamente al metrics server en lugar de usar el proxy de ArgoCD.

### Archivos Modificados

1. **`extensions/resource-metrics/resource-metrics-extention/ui/src/Metrics/client.ts`**
   - Cambiado: `/extensions/metrics/api/applications/...`
   - Por: `http://argocd-metrics-dev.devops.ligo.live/api/applications/...`

2. **`extensions/resource-metrics/resource-metrics-extention/ui/src/Metrics/Metrics.tsx`**
   - Cambiado: `/extensions/metrics/api/applications/...`
   - Por: `http://argocd-metrics-dev.devops.ligo.live/api/applications/...`

## Ventajas

1. **No requiere proxy de ArgoCD**: La extensión se conecta directamente al metrics server
2. **Headers siempre presentes**: La extensión UI construye los headers correctamente (`Argocd-Application-Name` y `Argocd-Project-Name`)
3. **Evita problemas con ALB**: No hay problemas de headers perdidos en el ALB

## Consideraciones

### CORS (Cross-Origin Resource Sharing)

Como la extensión UI se ejecuta en el navegador y hace peticiones a un dominio diferente (`argocd-metrics-dev.devops.ligo.live`), el metrics server debe tener CORS configurado.

El metrics server necesita responder con headers CORS como:
```
Access-Control-Allow-Origin: https://argocd-dev.devops.ligo.live
Access-Control-Allow-Methods: GET, POST, OPTIONS
Access-Control-Allow-Headers: Argocd-Application-Name, Argocd-Project-Name, Content-Type
```

### Verificar CORS en el Metrics Server

Si el metrics server no tiene CORS configurado, las peticiones fallarán con errores de CORS en el navegador.

**Solución**: Agregar middleware CORS al metrics server o configurar el ingress/ALB para agregar headers CORS.

## Construir la Extensión Modificada

### Prerequisitos

```bash
# Instalar Node.js y Yarn
node --version  # >= 14
yarn --version  # >= 1.22
```

### Pasos para Construir

**Opción 1: Usar el Makefile (Recomendado)**

Desde el directorio raíz del proyecto:
```bash
make build-ui
```

Esto:
1. Limpia builds anteriores
2. Instala dependencias con yarn
3. Construye la extensión
4. Crea el archivo `extension.tar` en `extensions/resource-metrics/resource-metrics-extention/ui/`

**Opción 2: Construcción Manual**

1. **Navegar al directorio de la UI**:
   ```bash
   cd extensions/resource-metrics/resource-metrics-extention/ui
   ```

2. **Instalar dependencias**:
   ```bash
   yarn install
   ```

3. **Construir la extensión**:
   ```bash
   yarn build
   ```

   Esto ejecuta: `webpack --config ./webpack.config.js && tar -C dist -cvf extension.tar resources`
   
   El archivo `extension.tar` se creará en el directorio `ui/`

### Crear el archivo extension.tar.gz

Si necesitas el formato `.tar.gz` en lugar de `.tar`:

```bash
# Desde el directorio ui/
cd extensions/resource-metrics/resource-metrics-extention/ui
gzip extension.tar
# Esto crea extension.tar.gz
```

O crear directamente:
```bash
cd extensions/resource-metrics/resource-metrics-extention/ui/dist
tar -czf ../extension.tar.gz resources
```

## Desplegar la Extensión Modificada

### Opción 1: Usar extension-installer (Recomendado)

Modificar el deployment de ArgoCD para usar el archivo local o un URL personalizado:

```yaml
initContainers:
  - name: extension-metrics
    image: quay.io/argoprojlabs/argocd-extension-installer:v0.0.1
    env:
    - name: EXTENSION_URL
      value: file:///path/to/extension.tar.gz  # O URL personalizada
    - name: EXTENSION_CHECKSUM_URL
      value: file:///path/to/extension_checksums.txt
    volumeMounts:
      - name: extensions
        mountPath: /tmp/extensions/
```

### Opción 2: Montar directamente

Si tienes acceso al filesystem del pod de ArgoCD, puedes montar el directorio directamente.

## Verificar que Funciona

1. **Abrir la consola del navegador** (F12)
2. **Ir a la pestaña Network**
3. **Cargar una aplicación en ArgoCD**
4. **Verificar que las peticiones vayan a**:
   - `http://argocd-metrics-dev.devops.ligo.live/api/applications/...`
   - En lugar de `/extensions/metrics/api/applications/...`

5. **Verificar headers en las peticiones**:
   - Debe incluir `Argocd-Application-Name: namespace:application-name`
   - Debe incluir `Argocd-Project-Name: project-name`

## Troubleshooting

### Error de CORS

Si ves errores como:
```
Access to fetch at 'http://argocd-metrics-dev.devops.ligo.live/...' from origin 'https://argocd-dev.devops.ligo.live' has been blocked by CORS policy
```

**Solución**: Configurar CORS en el metrics server o en el ALB/Ingress.

### Error 400: Headers faltantes

Si el metrics server devuelve error 400 sobre headers faltantes, verificar que:
1. La extensión esté construida con los cambios
2. Los headers se estén enviando correctamente (verificar en Network tab)

### La extensión no carga

1. Verificar que el archivo `extension.tar.gz` esté correctamente construido
2. Verificar logs del init container `extension-metrics` en el pod de ArgoCD
3. Verificar que el archivo esté en `/tmp/extensions/` dentro del pod

## Revertir Cambios

Si necesitas volver a usar el proxy de ArgoCD, simplemente revierte los cambios en los dos archivos modificados y reconstruye la extensión.
