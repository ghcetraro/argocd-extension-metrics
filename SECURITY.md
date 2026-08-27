# Seguridad

## Secretos

**Nunca commitees**:

- Kubeconfig, tokens de Argo CD o Kubernetes
- Usuario/clave o bearer de Prometheus en ConfigMaps versionados
- `.env` con credenciales reales

## Prometheus

1. Preferí Secret + env: `PROMETHEUS_USERNAME` / `PROMETHEUS_PASSWORD` o `PROMETHEUS_BEARER_TOKEN`
2. Alternativa: `password_file` / `bearer_token_file` montados desde Secret
3. Plantilla: `manifests/secret.example.yaml` (sin valores reales)
4. Detalle: [docs/configuracion-prometheus.md](docs/configuracion-prometheus.md)

## Buenas prácticas

1. RBAC mínimo para el ServiceAccount del metrics-server
2. HTTPS hacia Prometheus cuando esté en Internet
3. En producción, preferí `ca_file` frente a `insecure_skip_verify`
4. No loguear passwords ni tokens

## Reporte de vulnerabilidades

Contactá al mantenedor o usá [GitHub Security Advisories](https://github.com/ghcetraro/argocd-extension-metrics/security/advisories/new).
