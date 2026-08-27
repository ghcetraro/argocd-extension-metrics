# argocd-extension-metrics

Backend y extensión UI para visualizar métricas de recursos de Kubernetes directamente desde la interfaz de Argo CD, consultando Prometheus.

Basado en [argoproj-labs/argocd-extension-metrics](https://github.com/argoproj-labs/argocd-extension-metrics).

## Características

- Servidor HTTP (`argocd-metrics-server`) que expone queries a Prometheus
- Extensión UI embebida en Argo CD para gráficos de CPU, memoria y red por pod
- Manifiestos Kubernetes listos para desplegar
- Prometheus remoto por **HTTPS** con **basic auth** o **bearer token**
- Build de imagen Docker y UI con Makefile

## Requisitos

- Cluster Kubernetes con Argo CD instalado
- Prometheus accesible desde el cluster (HTTP/HTTPS; con o sin auth)
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

El Deployment expone el puerto `9003` (`GET /healthz`). La imagen por defecto es `quay.io/argoprojlabs/argocd-extension-metrics:latest`.

### Configurar Prometheus (HTTPS + clave)

Ver guía completa: [docs/configuracion-prometheus.md](docs/configuracion-prometheus.md).

Resumen:

1. En el ConfigMap, `provider.address`: `https://tu-prometheus…`
2. Crear Secret (plantilla: `manifests/secret.example.yaml`)
3. Descomentar env `PROMETHEUS_USERNAME` / `PROMETHEUS_PASSWORD` (o `PROMETHEUS_BEARER_TOKEN`) en el Deployment

### Configurar extensión en Argo CD

Registrar la extensión en el ConfigMap de Argo CD apuntando al servicio `argocd-metrics-server` y al bundle UI generado con `make build-ui`.

## Estructura del repositorio

| Carpeta / archivo | Descripción |
|-------------------|-------------|
| `cmd/` | Entrypoint del servidor |
| `internal/` | Lógica del servidor y cliente Prometheus |
| `extensions/` | Código de la extensión UI |
| `manifests/` | Deployment, Service, ConfigMap, ejemplo de Secret |
| `app/` | Config de desarrollo local |
| `docs/` | Documentación adicional |
| `REQUIREMENTS.md` | Alcance del proyecto |
| `.cursor/` | Reglas/hooks del agente Cursor |

## Documentación

- [Configurar Prometheus HTTPS/auth](docs/configuracion-prometheus.md)
- [Requerimientos](REQUIREMENTS.md)
- [Changelog](CHANGELOG.md)
- [Contribuir](CONTRIBUTING.md)
- [Seguridad](SECURITY.md)

## Licencia

Apache License 2.0 (ver [LICENSE](LICENSE)).
