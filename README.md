# argocd-extension-metrics

Backend y extensión UI para visualizar métricas de recursos de Kubernetes directamente desde la interfaz de ArgoCD, consultando Prometheus.

Basado en [argoproj-labs/argocd-extension-metrics](https://github.com/argoproj-labs/argocd-extension-metrics).

## Características

- Servidor HTTP (`argocd-metrics-server`) que expone queries a Prometheus
- Extensión UI embebida en ArgoCD para gráficos de CPU, memoria y red por pod
- Manifiestos Kubernetes listos para desplegar
- Build de imagen Docker y UI con Makefile

## Requisitos

- Cluster Kubernetes con ArgoCD instalado
- Prometheus accesible desde el cluster (o endpoint configurado en `config.json`)
- Go 1.21+ (para compilar)
- Node.js / Yarn (para compilar la UI)

## Inicio rápido

### Compilar

```bash
make build          # binario Go → dist/argocd-metrics-server
make build-ui       # extensión UI → extensions/.../ui/dist/
make image          # imagen Docker
make test           # tests unitarios
```

### Desplegar

```bash
kubectl apply -k manifests/
```

El Deployment expone el puerto `9003`. La imagen por defecto es `quay.io/argoprojlabs/argocd-extension-metrics:latest`.

### Configurar extensión en ArgoCD

Registrar la extensión en el ConfigMap de ArgoCD apuntando al servicio `argocd-metrics-server` y al bundle UI generado con `make build-ui`.

## Estructura del repositorio

| Carpeta | Descripción |
|---------|-------------|
| `cmd/` | Entrypoint del servidor |
| `internal/` | Lógica del servidor y queries Prometheus |
| `extensions/` | Código de la extensión UI |
| `manifests/` | Deployment, Service y ConfigMap |
| `app/` | Configuración de la aplicación |
| `docs/` | Documentación adicional |

## Variables de entorno / configuración

La configuración de Prometheus y las queries se define en el ConfigMap (`manifests/configmap.yaml`) montado como `/app/config.json`.

## Licencia

Ver [LICENSE](LICENSE).
