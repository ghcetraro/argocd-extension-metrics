# Seguridad

## Secretos

**Nunca commitees**:

- Kubeconfig, tokens de ArgoCD o Kubernetes
- Configuración con endpoints/credenciales de Prometheus en producción

## Buenas prácticas

1. Montar configuración sensible via Secrets/ConfigMaps en el cluster
2. Restringir RBAC del ServiceAccount al mínimo necesario
3. TLS en tráfico externo cuando sea posible

## Consideraciones del proyecto

- Configuración de Prometheus con credenciales en ConfigMap/Secret
- Acceso al API server de Kubernetes desde el pod del metrics-server

## Reporte de vulnerabilidades

Contactá al mantenedor o usá [GitHub Security Advisories](https://github.com/ghcetraro/argocd-extension-metrics/security/advisories/new).
