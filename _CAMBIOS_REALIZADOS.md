# Cambios Realizados en el Código del Metrics Server

## Problema Identificado

El template `{{.name}}` no se estaba reemplazando, resultando en queries con `pod=~""` (vacío) y errores 400 con `{}`.

## Cambios Aplicados

### 1. Cambio de `html/template` a `text/template`

**Archivo:** `internal/server/prometheus.go` línea 7

**Antes:**
```go
import "html/template"
```

**Después:**
```go
import "text/template"
```

**Razón:** 
- `html/template` hace auto-escape de caracteres, lo que puede interferir con queries de Prometheus
- `text/template` es más simple y adecuado para queries
- Consistente con el test (`prometheus_test.go`)

### 2. Mejora del Manejo de Errores

**Archivo:** `internal/server/prometheus.go` líneas 158-160, 163-165, 181-184, 186-194, 201-204

**Antes:**
```go
if err != nil {
    ctx.JSON(http.StatusBadRequest, err)  // ← No se serializa bien
    return
}
```

**Después:**
```go
if err != nil {
    pp.logger.Errorf("Error executing graph query: %v", err)  // ← Logging
    ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})  // ← Serialización correcta
    return
}
```

**Razón:** 
- Los errores ahora se serializan correctamente a JSON
- Se agrega logging para diagnóstico

### 3. Agregado de Logging para Diagnóstico

**Archivo:** `internal/server/prometheus.go` líneas 88-99

**Agregado:**
```go
// Logging para diagnóstico
pp.logger.Infof("Template queryExpression: %s", queryExpression)
pp.logger.Infof("Template env (raw): %+v", env)
pp.logger.Infof("Template env1 (processed): %+v", env1)
if nameVal, ok := env1["name"]; ok {
    pp.logger.Infof("Template env1['name'] value: '%s'", nameVal)
} else {
    pp.logger.Warnf("Template env1['name'] NOT FOUND in map")
}

// ... ejecutar template ...

pp.logger.Infof("Final query after template execution: %s", strQuery)
```

**Razón:** 
- Permite ver qué query params están llegando
- Permite ver qué query se genera después del template
- Facilita el diagnóstico de problemas

## Próximos Pasos

1. **Rebuild del binario:**
   ```bash
   cd argocd-extension-metrics
   make build
   ```

2. **Crear nueva imagen Docker** (si es necesario)

3. **Actualizar el deployment** para usar la nueva imagen

4. **Probar el curl** y revisar los logs:
   ```bash
   kubectl logs -f -n argocd deployment/argocd-metrics-server
   ```

5. **Verificar que el template funcione** y que los logs muestren:
   - Los query params recibidos
   - El valor de `name`
   - La query final después del template

## Verificación

Después del rebuild y redeploy, los logs deberían mostrar:
```
Template queryExpression: sum(rate(container_cpu_usage_seconds_total{pod=~"{{.name}}", ...}[5m])) by (container)
Template env (raw): map[name:[core-auth-payfi-59f897bf6f-dpp2q.*] namespace:[payfi] ...]
Template env1 (processed): map[name:core-auth-payfi-59f897bf6f-dpp2q.* namespace:payfi ...]
Template env1['name'] value: 'core-auth-payfi-59f897bf6f-dpp2q.*'
Final query after template execution: sum(rate(container_cpu_usage_seconds_total{pod=~"core-auth-payfi-59f897bf6f-dpp2q.*", ...}[5m])) by (container)
```

Si el template funciona, la query debería ejecutarse correctamente en Prometheus.
