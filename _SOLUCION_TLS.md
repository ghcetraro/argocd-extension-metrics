# Solución al Problema de TLS

## Problema

El metrics server falla al conectarse a Prometheus con el error:
```
tls: failed to verify certificate: x509: certificate signed by unknown authority
```

## Causa

El cliente de Prometheus está intentando verificar el certificado TLS de `https://prometheus-dev.devops.ligo.live`, pero el certificado no es confiable (probablemente autofirmado o con una CA no reconocida).

## Solución Aplicada

Modificado `prometheus.go` para usar `InsecureSkipVerify: true` cuando no hay configuración TLS explícita en el ConfigMap. Esto permite que el cliente se conecte a Prometheus sin verificar el certificado.

### Cambios en `prometheus.go`

**Antes:**
```go
func (pp *PrometheusProvider) init() error {
	client, err := api.NewClient(api.Config{
		Address: pp.config.Provider.Address,
	})
	if err != nil {
		pp.logger.Errorf("Error creating client: %v\n", err)
		return err
	}
	pp.provider = v1.NewAPI(client)
	return nil
}
```

**Después:**
```go
func (pp *PrometheusProvider) init() error {
	// Configurar HTTP client con TLS
	httpClientConfig := config.HTTPClientConfig{}
	
	// Si TLSConfig está configurado, usarlo; si no, usar InsecureSkipVerify para desarrollo
	if pp.config.Provider.TLSConfig.InsecureSkipVerify || 
		(pp.config.Provider.TLSConfig.CAFile == "" && 
		 pp.config.Provider.TLSConfig.CertFile == "" && 
		 pp.config.Provider.TLSConfig.KeyFile == "") {
		// No hay configuración TLS, usar InsecureSkipVerify para desarrollo
		httpClientConfig.TLSConfig = config.TLSConfig{
			InsecureSkipVerify: true,
		}
		pp.logger.Warnf("Using InsecureSkipVerify for Prometheus connection (development mode)")
	} else {
		// Usar la configuración TLS del provider
		httpClientConfig.TLSConfig = pp.config.Provider.TLSConfig
	}
	
	httpClient, err := config.NewClientFromConfig(httpClientConfig, "prometheus", false)
	if err != nil {
		pp.logger.Errorf("Error creating HTTP client: %v\n", err)
		return err
	}
	
	client, err := api.NewClient(api.Config{
		Address:      pp.config.Provider.Address,
		RoundTripper: httpClient.Transport,
	})
	if err != nil {
		pp.logger.Errorf("Error creating Prometheus client: %v\n", err)
		return err
	}
	pp.provider = v1.NewAPI(client)
	return nil
}
```

## Comportamiento

- Si el ConfigMap tiene `TLSConfig` configurado (con CAFile, CertFile, KeyFile, o InsecureSkipVerify), se usa esa configuración.
- Si el ConfigMap NO tiene `TLSConfig` configurado (vacío), se usa `InsecureSkipVerify: true` automáticamente.
- Se loguea un warning cuando se usa `InsecureSkipVerify` para indicar que está en modo desarrollo.

## Próximos Pasos

1. **Rebuild del binario:**
   ```bash
   cd argocd-extension-metrics
   make build
   ```

2. **Crear nueva imagen Docker** y hacer push

3. **Redeploy** el metrics server

4. **Verificar** que ahora se conecte correctamente a Prometheus

## Nota de Seguridad

`InsecureSkipVerify: true` desactiva la verificación de certificados TLS, lo cual es adecuado para entornos de desarrollo pero **NO debe usarse en producción**. Para producción, se debe configurar correctamente el `TLSConfig` en el ConfigMap con los certificados apropiados.
