# Changelog

Formato basado en [Keep a Changelog](https://keepachangelog.com/es-ES/1.1.0/).
Versionado según [Semantic Versioning](https://semver.org/lang/es/).

## [Unreleased]

### Added

- Autenticación hacia Prometheus: basic auth y bearer token vía env (`PROMETHEUS_*`) o ConfigMap (`basic_auth`, `bearer_token`, `bearer_token_file`)
- Override de URL con `PROMETHEUS_ADDRESS` (prioridad sobre el ConfigMap)
- Normalización de URL Grafana Cloud (`…grafana.net` → `…/api/prom`)
- Ejemplo de Secret Kubernetes (`manifests/secret.example.yaml`) y documentación en `docs/configuracion-prometheus.md`
- Probes `/healthz` y `revisionHistoryLimit: 0` en el Deployment
- `REQUIREMENTS.md`, `WHITELIST.md` y reglas Cursor adaptadas (`.cursor/`)

## [1.0.0] - 2026-07-04

Primera release pública.

### Added

- Servidor de métricas Prometheus para extensión de ArgoCD
- Extensión UI para gráficos de recursos en pods
- Manifiestos Kubernetes (Deployment, Service, ConfigMap)
- Makefile para build, tests e imagen Docker
- Documentación de despliegue en README.md

[1.0.0]: https://github.com/ghcetraro/argocd-extension-metrics/releases/tag/v1.0.0
