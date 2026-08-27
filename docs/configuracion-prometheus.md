# Configurar Prometheus (HTTPS + autenticación)

El metrics-server lee `app/config.json` (ConfigMap) y opcionalmente variables de entorno para autenticarse contra Prometheus.

## URL HTTPS pública

En el ConfigMap, bloque `prometheus.provider`:

```json
"provider": {
  "name": "default",
  "default": true,
  "address": "https://prometheus.ejemplo.com",
  "TLSConfig": {
    "insecure_skip_verify": true
  }
}
```

- Sin `ca_file` / `cert_file` / `key_file`, el servidor usa `InsecureSkipVerify` (útil con cert autofirmado).
- Producción con CA propia: montá el CA en el pod y usá `"ca_file": "/ruta/ca.crt"` + `"insecure_skip_verify": false`.

## Autenticación (basic o bearer)

**Preferido:** Secret de Kubernetes + env en el Deployment (ver `manifests/secret.example.yaml`).

| Variable | Uso |
|----------|-----|
| `PROMETHEUS_USERNAME` | Usuario basic auth |
| `PROMETHEUS_PASSWORD` | Clave basic auth |
| `PROMETHEUS_PASSWORD_FILE` | Archivo con la clave (alternativa) |
| `PROMETHEUS_BEARER_TOKEN` | Token Bearer |
| `PROMETHEUS_BEARER_TOKEN_FILE` | Archivo con el token |

Las env **tienen prioridad** sobre el ConfigMap.

### Ejemplo: basic auth vía Secret

```bash
kubectl create secret generic argocd-metrics-prometheus-auth \
  --from-literal=username='usuario' \
  --from-literal=password='clave' \
  -n <namespace>
```

Descomentá el bloque `env` en `manifests/deployment.yaml` y aplicá.

### Alternativa en ConfigMap (menos seguro)

```json
"provider": {
  "name": "default",
  "default": true,
  "address": "https://prometheus.ejemplo.com",
  "basic_auth": {
    "username": "usuario",
    "password_file": "/etc/prometheus/password"
  },
  "TLSConfig": {
    "insecure_skip_verify": true
  }
}
```

O bearer:

```json
"bearer_token_file": "/etc/prometheus/token"
```

No uses basic y bearer a la vez.

## Health check

- `GET /healthz` → `200` / `healthy` (sin auth)
- Puerto del contenedor: `9003`

## Checklist

1. Pod del metrics-server con egress a la URL pública
2. ConfigMap con `address` `https://…`
3. Secret + env (o `*_file` montados)
4. Reiniciar el Deployment tras cambiar config/secret
