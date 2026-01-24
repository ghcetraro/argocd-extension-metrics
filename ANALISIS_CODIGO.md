# Análisis del Código de argocd-extension-metrics

## Validación de Headers en `server.go`

El servidor valida estrictamente los headers HTTP. Aquí está el flujo de validación:

### 1. Validación de Headers Requeridos

```go
// Línea 147-151: Valida que exista Argocd-Application-Name
if err := validateHeader(headers, "Argocd-Application-Name"); err != nil {
    ms.logger.Warn(err)
    ctx.JSON(400, gin.H{"error": err.Error()})
    return
}

// Línea 152-153: Extrae el application name del header
val := headers["Argocd-Application-Name"]
applicationNameHeader := strings.Split(val[0], ":")[1]  // ⚠️ Divide por ":" y toma la segunda parte
```

**IMPORTANTE**: El header `Argocd-Application-Name` debe tener formato: `namespace:application-name`

```go
// Línea 155-161: Valida que exista Argocd-Project-Name
if err := validateHeader(headers, "Argocd-Project-Name"); err != nil {
    ms.logger.Warn(err)
    ctx.JSON(400, gin.H{"error": err.Error()})
    return
}
temp := headers["Argocd-Project-Name"]
projectHeader := temp[0]
```

### 2. Validación de Query Parameters

```go
// Línea 163-169: Valida query param application_name
applicationNameQueryParam := ctx.Query("application_name")
if err := validateQueryParam(applicationNameQueryParam, "application_name"); err != nil {
    // Error 400
}

// Línea 171-177: Valida query param project
projectQueryParam := ctx.Query("project")
if err := validateQueryParam(projectQueryParam, "project"); err != nil {
    // Error 400
}
```

### 3. Validación de Coincidencia

```go
// Línea 179-185: El application_name del header DEBE coincidir con el query param
if applicationNameHeader != applicationNameQueryParam {
    msg := "Application name mismatch. Value from the header is different from the url."
    ctx.JSON(400, gin.H{"error": err.Error()})
    return
}

// Línea 187-193: El project del header DEBE coincidir con el query param
if projectHeader != projectQueryParam {
    msg := "Project mismatch. Value from the header is different from the url."
    ctx.JSON(400, gin.H{"error": err.Error()})
    return
}
```

## Requisitos Estrictos

1. **Header `Argocd-Application-Name`**:
   - Debe existir
   - Formato: `namespace:application-name` (debe tener ":" y el código toma la segunda parte)
   - El `application-name` (después de ":") debe coincidir exactamente con el query param `application_name`

2. **Header `Argocd-Project-Name`**:
   - Debe existir
   - Debe coincidir exactamente con el query param `project`

3. **Query Parameters**:
   - `application_name`: Requerido, debe coincidir con la segunda parte del header
   - `project`: Requerido, debe coincidir con el header

## Ejemplo Correcto

```bash
# Headers
Argocd-Application-Name: payfi:core-auth-payfi-development
Argocd-Project-Name: development

# Query params
?application_name=core-auth-payfi-development&project=development

# El código extrae "core-auth-payfi-development" del header (después de ":")
# y lo compara con application_name del query param
# Ambos deben ser iguales
```

## Problema Identificado

El código fuente confirma que:
1. Los headers son **obligatorios** y no hay forma de deshabilitar esta validación
2. El formato del header `Argocd-Application-Name` debe ser `namespace:application-name`
3. Los valores deben coincidir exactamente con los query params

Por lo tanto, el proxy sidecar que creamos es la solución correcta, pero debe:
1. Construir el header en formato `namespace:application-name`
2. Asegurarse de que los valores coincidan con los query params

## Función validateHeader

```go
func validateHeader(header http.Header, headerName string) error {
    val, ok := header[headerName]
    if !ok {
        errMsg := headerName + " header not sent"
        return errors.New(errMsg)
    }
    if len(val) != 1 {
        errMsg := "Multiple values for " + headerName + " header sent. Only one is allowed"
        return errors.New(errMsg)
    }
    return nil
}
```

Esta función valida que:
- El header exista
- Solo haya un valor (no múltiples valores)

## Conclusión

El código fuente confirma que el problema es que ArgoCD no está enviando los headers cuando hace proxy. La solución del proxy sidecar es correcta, pero debe asegurarse de:

1. Construir `Argocd-Application-Name` en formato `namespace:application-name`
2. Construir `Argocd-Project-Name` desde el query param `project`
3. Asegurarse de que los valores coincidan exactamente con los query params
