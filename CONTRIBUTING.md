# Contribuir a argocd-extension-metrics

Gracias por interesarte en el proyecto.

## Antes de empezar

1. Revisá [README.md](README.md).
2. **No commitees secretos**: kubeconfig, tokens, `.env`.

## Cómo reportar bugs

1. Buscá si ya existe un [issue](https://github.com/ghcetraro/argocd-extension-metrics/issues) similar.
2. Abrí uno nuevo con:
   - Versión / commit
   - Configuración relevante (**sin secretos**)
   - Comportamiento esperado vs actual

Para vulnerabilidades, seguí [SECURITY.md](SECURITY.md).

## Pull requests

1. Fork del repo y branch desde `master` (o `main` si es el default del remoto):
   ```bash
   git checkout -b feature/mi-cambio
   ```
2. Cambios acotados y commits claros en español o inglés.
3. Verificá localmente:
   ```bash
   make test
   make build
   ```
4. Actualizá README.md, `docs/` y CHANGELOG.md si cambiás comportamiento.
5. Abrí el PR describiendo el **por qué** del cambio.

## Estilo

- Go 1.21+, seguir estilo del código existente
- Health: `GET /healthz` sin auth
- Credenciales Prometheus solo por Secret/env, nunca en git
- Cambios mínimos por PR

## Releases

Versiones etiquetadas (`v1.0.0`, …) documentadas en [CHANGELOG.md](CHANGELOG.md).

## Preguntas

Abrí un issue con etiqueta `question`.
