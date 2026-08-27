# Requerimientos — argocd-extension-metrics

## Propósito

Servidor de métricas y extensión UI para Argo CD que consulta Prometheus y muestra gráficos de recursos (CPU, memoria, etc.) en la UI de Argo CD.

## Alcance

- Debe hacer:
  - Exponer API HTTP para dashboards/queries Prometheus
  - Extensión UI embebible en Argo CD
  - Conectarse a Prometheus por HTTP/HTTPS, en cluster o en Internet
  - Autenticación hacia Prometheus: basic auth y bearer token (env o config)
  - TLS configurable (`TLSConfig` / skip-verify cuando no hay CA)
- No debe hacer (no-goals):
  - Sustituir Grafana u otros backends de observabilidad
  - Gestionar Prometheus ni scrapes
  - Autenticar usuarios finales de Argo CD (eso lo hace Argo CD)

## Cómo tratar este proyecto

- Prioridad: estabilidad de la conexión a Prometheus y compatibilidad con el ConfigMap existente
- Credenciales nunca en git: Secret de Kubernetes + env (`PROMETHEUS_*`)
- Cambios mínimos; no refactors no pedidos
- Documentar en español; código/comentarios pueden mezclar EN/ES según lo existente
- Branch habitual: `master` / `main` (público en GitHub)

## Criterios de aceptación (auth HTTPS)

- Con `address: https://…` el cliente habla HTTPS
- Con `PROMETHEUS_USERNAME` + `PROMETHEUS_PASSWORD` (o `basic_auth` en config) envía basic auth
- Con `PROMETHEUS_BEARER_TOKEN` (o `bearer_token` / `bearer_token_file`) envía Bearer
- Basic y bearer son mutuamente excluyentes
