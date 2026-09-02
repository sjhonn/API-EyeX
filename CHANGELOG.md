# Changelog

Todos los cambios relevantes de EyeX se documentan en este archivo. El proyecto sigue Semantic Versioning.

## [1.2.0] - 2026-09-01

### Added

- `POST /api/v1/simulate` para simulación de un color.
- `POST /api/v1/simulate/batch` para lotes de 1 a 256 colores.
- Matrices de Machado, Oliveira & Fernandes (2009) para protanopia, deuteranopia y tritanopia.
- Severidad continua `0..1` con interpolación lineal entre matrices tabuladas.
- Paridad de simulación en Go, PHP, TypeScript/Node y Java.
- Vectores de referencia y pruebas HTTP de paridad en CI.
- Métodos `simulate` y `simulateBatch` en React, React Native, Flutter y SDK TypeScript genérico.
- README independiente para cada SDK principal.
- Quickstart, badges, política SemVer/deprecación y documentación detallada de extensión.
- `CONTRIBUTING.md`, `CHANGELOG.md` y `LICENSE`.
- Métricas Prometheus, rate limit, API key opcional, timeout, security headers, HTTPS de producción y logs JSON en el servidor Go.
- ETag/Cache-Control para temas y exposición de `openapi.yaml` desde el servidor Go.

### Changed

- `openapi.yaml` actualizado a `1.2.0` de manera aditiva.
- Versiones de SDKs y backends de referencia alineadas con `1.2.0`.
- Documentación de alcance: EyeX cubre color, no sustituye navegación por teclado, ARIA ni una auditoría WCAG completa.
- CI ampliado con cobertura Go, validación OpenAPI, build de clientes web y paridad real de los cuatro runtimes.
- Playgrounds HTML/JavaScript y Vue actualizados para demostrar la simulación; demo `web/` retirada de endpoints obsoletos y alineada con API v1.2.

### Compatibility

- Las rutas de tema de `/api/v1/` permanecen sin eliminación ni renombre.
- La simulación se agrega como capacidad nueva; no cambia el formato de los endpoints existentes.

## [1.1.0] - 2026-08-27

### Added

- Temas `normal` y `low_vision`, modos claro/oscuro e intensidad configurable.
- Alto contraste y validación de contraste para texto.
- `POST /api/v1/theme/custom`.
- `POST /api/v1/test/suggest`.
- SDKs React, React Native y Flutter.
- Extensión de navegador, CSS estático, widget embebible y cliente Vue.
- Ampliación de la equivalencia entre Go, PHP, TypeScript/Node y Java.

## [1.0.0] - 2026-08-25

### Added

- Primera versión pública de EyeX.
- `GET /api/v1/theme/types` y `GET /api/v1/theme/{type}`.
- Backend principal Go y backends equivalentes PHP, TypeScript/Node y Java.
- Playground HTML/JavaScript, Dockerfile y manifiestos de infraestructura.
- Utilidades iniciales de color y pruebas Go.
