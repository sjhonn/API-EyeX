# EyeX

[![EyeX CI](https://github.com/sjhonn/API-EyeX/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/sjhonn/API-EyeX/actions/workflows/ci.yml)
[![Coverage](https://img.shields.io/badge/coverage-63.9%25-yellowgreen)](https://github.com/sjhonn/API-EyeX/actions/workflows/ci.yml)
[![Go](https://img.shields.io/badge/Go-1.23%2B-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

<img src="frontend/html-js/assets/logo.png" alt="Logo de EyeX" width="180">

EyeX es una API de accesibilidad visual orientada a color. Entrega temas para daltonismo, baja visión y alto contraste, y desde `v1.2.0` también permite simular cómo cambia un color bajo protanopia, deuteranopia o tritanopia con severidad continua.

El contrato HTTP principal está disponible en Go y mantiene implementaciones equivalentes en PHP, TypeScript/Node y Java. El repositorio también incluye clientes web, SDKs, una extensión de navegador y manifiestos de infraestructura. Este README se concentra en desarrollo y uso del API; el hosting no forma parte del alcance de esta versión.

## Quickstart — 30 segundos

Requisito: Go 1.23 o superior.

```bash
go run ./cmd/api
```

En otra terminal:

```bash
curl http://localhost:8080/api/v1/theme/types

curl -X POST http://localhost:8080/api/v1/simulate \
  -H 'Content-Type: application/json' \
  -d '{"hex":"#FF0000","type":"protanopia","severity":0.65}'
```

Resultado de la simulación:

```json
{
  "original": "#FF0000",
  "simulated": "#A05A00",
  "type": "protanopia",
  "severity": 0.65,
  "model": "machado-2009"
}
```

La web principal queda disponible en `http://localhost:8080/` y el contrato OpenAPI en `http://localhost:8080/openapi.yaml`.

## Funcionalidad

EyeX incluye:

- temas para `normal`, `protanopia`, `deuteranopia`, `tritanopia`, `achromatopsia` y `low_vision`;
- intensidad de temas `mild`, `moderate` o `severe`;
- modos `dark` y `light`;
- opción de alto contraste;
- validación WCAG AA de texto contra `background` y `surface` con umbral 4.5:1;
- adaptación de paletas propias con `POST /api/v1/theme/custom`;
- test orientativo con `POST /api/v1/test/suggest`;
- simulación de color individual con `POST /api/v1/simulate`;
- simulación de hasta 256 colores con `POST /api/v1/simulate/batch`;
- implementación de simulación equivalente en Go, PHP, TypeScript/Node y Java;
- OpenAPI 3.0.3, CI, pruebas de paridad, cobertura Go, SDKs y extensión de navegador;
- CORS, JSON logging, ETag para temas, rate limit, API key opcional, headers de seguridad, timeout y métricas Prometheus en el servidor Go oficial.

## Endpoints del API v1

| Método | Ruta | Función |
| --- | --- | --- |
| `GET` | `/api/v1/theme/types` | Lista tipos de tema. |
| `GET` | `/api/v1/theme/{type}` | Devuelve una paleta de tema. |
| `POST` | `/api/v1/theme/custom` | Adapta una paleta proporcionada por el cliente. |
| `POST` | `/api/v1/test/suggest` | Sugiere un tipo de tema de forma orientativa. |
| `POST` | `/api/v1/simulate` | Simula un color con severidad continua. |
| `POST` | `/api/v1/simulate/batch` | Simula entre 1 y 256 colores. |
| `GET` | `/metrics` | Métricas Prometheus del servidor Go. |
| `GET` | `/openapi.yaml` | Contrato OpenAPI del servidor Go. |
| `GET` | `/eyex.css` | Paletas CSS estáticas. |
| `GET` | `/eyex.js` | Widget embebible. |

### Tipos de tema

```bash
curl http://localhost:8080/api/v1/theme/types
```

```json
{
  "types": [
    "normal",
    "protanopia",
    "deuteranopia",
    "tritanopia",
    "achromatopsia",
    "low_vision"
  ]
}
```

### Obtener un tema

```bash
curl 'http://localhost:8080/api/v1/theme/deuteranopia?severity=moderate&mode=dark&high_contrast=true'
```

Parámetros opcionales:

| Parámetro | Valores | Uso |
| --- | --- | --- |
| `severity` | `mild`, `moderate`, `severe` | Intensidad de adaptación del tema. |
| `mode` | `dark`, `light` | Variante clara u oscura. |
| `high_contrast` | `true`, `false` | Refuerza el contraste. |

Sin parámetros se conserva el comportamiento compatible con versiones anteriores.

### Adaptar una paleta propia

```bash
curl -X POST http://localhost:8080/api/v1/theme/custom \
  -H 'Content-Type: application/json' \
  -d '{
    "type":"deuteranopia",
    "severity":"moderate",
    "mode":"dark",
    "high_contrast":true,
    "palette":{
      "background":"#101820",
      "surface":"#182430",
      "text":"#F8F9FA",
      "primary":"#E63946",
      "secondary":"#2A9D8F",
      "error":"#D62828",
      "success":"#2A9D8F"
    }
  }'
```

Todos los colores de entrada deben usar `#RRGGBB`.

### Test orientativo

```bash
curl -X POST http://localhost:8080/api/v1/test/suggest \
  -H 'Content-Type: application/json' \
  -d '{
    "answers":{
      "reds_look_darker":false,
      "green_brown_confusion":false,
      "blue_yellow_confusion":true,
      "colors_look_gray":false
    }
  }'
```

El resultado es una sugerencia de configuración. No constituye un diagnóstico médico.

## Simulación de colores — v1.2.0

### `POST /api/v1/simulate`

Entrada:

```json
{
  "hex": "#FF0000",
  "type": "protanopia",
  "severity": 0.65
}
```

`type` admite exclusivamente:

- `protanopia`;
- `deuteranopia`;
- `tritanopia`.

`severity` es un número continuo entre `0` y `1`:

- `0`: identidad, no modifica el color;
- `1`: simulación completa de la deficiencia;
- valores intermedios: interpolación lineal de las matrices tabuladas cada `0.1`.

Si `severity` se omite, EyeX usa `1`.

Ejemplo:

```bash
curl -X POST http://localhost:8080/api/v1/simulate \
  -H 'Content-Type: application/json' \
  -d '{"hex":"#FF0000","type":"protanopia","severity":0.65}'
```

Respuesta:

```json
{
  "original": "#FF0000",
  "simulated": "#A05A00",
  "type": "protanopia",
  "severity": 0.65,
  "model": "machado-2009"
}
```

### `POST /api/v1/simulate/batch`

Aplica el mismo `type` y `severity` a una lista de entre 1 y 256 colores.

```bash
curl -X POST http://localhost:8080/api/v1/simulate/batch \
  -H 'Content-Type: application/json' \
  -d '{
    "colors":["#FF0000","#00FF00","#0000FF"],
    "type":"deuteranopia",
    "severity":0.5
  }'
```

Respuesta:

```json
{
  "type": "deuteranopia",
  "severity": 0.5,
  "model": "machado-2009",
  "results": [
    {"original":"#FF0000","simulated":"#C37600"},
    {"original":"#00FF00","simulated":"#CDE52E"},
    {"original":"#0000FF","simulated":"#0036FD"}
  ]
}
```

### Modelo matemático

La simulación usa las matrices publicadas por Gustavo M. Machado, Manuel M. Oliveira y Leandro A. F. Fernandes en *A Physiologically-based Model for Simulation of Color Vision Deficiency*, IEEE TVCG 15(6), 2009, DOI `10.1109/TVCG.2009.113`.

La implementación:

1. convierte sRGB con gamma a RGB lineal;
2. interpola las matrices de severidad cuando el valor no cae exactamente en un paso de `0.1`;
3. aplica la multiplicación matricial en RGB lineal;
4. limita cada canal al rango representable;
5. convierte nuevamente a sRGB y devuelve `#RRGGBB` en mayúsculas.

Los cuatro backends comparten los mismos coeficientes, reglas de interpolación, conversión sRGB y redondeo de salida.

## Paridad entre Go, PHP, TypeScript/Node y Java

[![Última paridad verificada](https://img.shields.io/github/actions/workflow/status/sjhonn/API-EyeX/ci.yml?branch=main&label=%C3%BAltima%20paridad%20verificada%20%E2%9C%85)](https://github.com/sjhonn/API-EyeX/actions/workflows/ci.yml?query=branch%3Amain)

El workflow `EyeX CI` levanta los cuatro backends en puertos independientes y ejecuta `tests/parity.py`. Se comparan status HTTP y JSON, y además se validan vectores conocidos de Machado para evitar que una misma regresión pase inadvertida en los cuatro lenguajes.

El enlace del badge abre los runs del workflow; el run exitoso más reciente de `main` representa la última paridad verificada en GitHub Actions. Los cambios locales de este ZIP quedan verificados por las pruebas descritas más abajo y deberán obtener su propio run al publicarse.

Casos cubiertos por la paridad HTTP:

- listado de tipos, tema legado y tema con parámetros;
- paleta personalizada y test orientativo;
- errores históricos `invalid_type` e `invalid_parameter`;
- severidades interpoladas de referencia (`0.25`, `0.5`, `0.65`) y los 11 anchors `0.0..1.0` de cada familia Machado;
- protanopia, deuteranopia y tritanopia;
- normalización de hexadecimal;
- severidad por defecto;
- batch;
- `invalid_type`;
- `invalid_parameter`;
- `invalid_color`;
- `invalid_request` para batch vacío, `colors` no-array y campos JSON no permitidos.

## Errores y ejemplos `curl`

Los errores controlados usan:

```json
{
  "error": "invalid_type",
  "message": "Tipo de daltonismo no soportado"
}
```

### `invalid_type` — HTTP 400

```bash
curl -i http://localhost:8080/api/v1/theme/no-existe
```

También se produce al usar un tipo no permitido en `/simulate`.

### `invalid_parameter` — HTTP 400

```bash
curl -i -X POST http://localhost:8080/api/v1/simulate \
  -H 'Content-Type: application/json' \
  -d '{"hex":"#FF0000","type":"protanopia","severity":1.5}'
```

### `invalid_request` — HTTP 400

```bash
curl -i -X POST http://localhost:8080/api/v1/simulate \
  -H 'Content-Type: application/json' \
  --data-binary '{json-invalido'
```

### `invalid_palette` — HTTP 400

```bash
curl -i -X POST http://localhost:8080/api/v1/theme/custom \
  -H 'Content-Type: application/json' \
  -d '{"type":"deuteranopia","palette":{"background":"red"}}'
```

### `invalid_color` — HTTP 400

```bash
curl -i -X POST http://localhost:8080/api/v1/simulate \
  -H 'Content-Type: application/json' \
  -d '{"hex":"red","type":"protanopia","severity":0.5}'
```

### `not_found` — HTTP 404

```bash
curl -i http://localhost:8080/api/v1/no-existe
```

### `method_not_allowed` — HTTP 405

```bash
curl -i -X DELETE http://localhost:8080/api/v1/theme/types
```

### `https_required` — HTTP 426, servidor Go en producción

Ejecuta temporalmente otra instancia:

```bash
EYEX_ENV=production EYEX_PORT=8089 go run ./cmd/api
```

Luego realiza una petición HTTP sin TLS ni `X-Forwarded-Proto: https`:

```bash
curl -i http://localhost:8089/api/v1/theme/types
```

### `rate_limited` — HTTP 429, servidor Go

Ejecuta temporalmente una instancia con límite de una petición por minuto:

```bash
EYEX_RATE_LIMIT_PER_MINUTE=1 EYEX_PORT=8089 go run ./cmd/api
```

En otra terminal:

```bash
curl -i http://localhost:8089/api/v1/theme/types
curl -i http://localhost:8089/api/v1/theme/types
```

La segunda petición responde `429` e incluye `Retry-After`.

### `timeout` — HTTP 504

Cualquier request puede recibirlo si el procesamiento supera `EYEX_REQUEST_TIMEOUT_MS`. Ejemplo de request sobre el que aplicaría el timeout configurado:

```bash
curl -i http://localhost:8080/api/v1/theme/deuteranopia
```

No se publica un endpoint artificialmente lento solo para forzar este error. El comportamiento del middleware se prueba de forma determinista en `internal/middleware/middleware_test.go`.

### `internal_server_error` — HTTP 500

Cualquier endpoint puede devolverlo si ocurre un fallo inesperado recuperado por el servidor. Ejemplo de request normal sujeto a esa protección:

```bash
curl -i http://localhost:8080/api/v1/theme/types
```

No se expone una ruta pública que provoque un panic deliberadamente. El servidor devuelve JSON sin stack trace al cliente y registra el incidente en los logs del proceso.

### Mensajes en inglés

El servidor Go utiliza `Accept-Language` en errores controlados:

```bash
curl -H 'Accept-Language: en' http://localhost:8080/api/v1/theme/no-existe
```

Respuesta:

```json
{
  "error": "invalid_type",
  "message": "Unsupported color vision type"
}
```

## Caché HTTP de temas

`GET /api/v1/theme/{type}` devuelve `ETag` y `Cache-Control`.

```bash
curl -i http://localhost:8080/api/v1/theme/deuteranopia
```

Reutiliza el ETag:

```bash
curl -i \
  -H 'If-None-Match: "ETAG_RECIBIDO"' \
  http://localhost:8080/api/v1/theme/deuteranopia
```

Si no cambió, el servidor Go responde `304 Not Modified`.

## Configuración

EyeX conserva un único archivo `.env` en la raíz. No se requieren ni se generan archivos `.env` adicionales para PHP, TypeScript/Node, Java, SDKs o frontends.

Variables soportadas por el servidor Go:

| Variable | Default | Uso |
| --- | --- | --- |
| `EYEX_PORT` | `8080` | Puerto HTTP. |
| `EYEX_ALLOWED_ORIGIN` | `*` | Origin permitido por CORS. |
| `EYEX_ENV` | `development` | `development`, `test` o `production`. |
| `EYEX_RATE_LIMIT_PER_MINUTE` | `60` | Solicitudes públicas por minuto e IP. |
| `EYEX_REQUEST_TIMEOUT_MS` | `5000` | Tiempo máximo por request. |
| `EYEX_API_KEY` | vacío | Si está configurada y coincide con `X-API-Key`, omite el rate limit público. |

## Seguridad y observabilidad del servidor Go

El backend Go añade:

- `Content-Security-Policy`;
- `X-Content-Type-Options: nosniff`;
- `X-Frame-Options: DENY`;
- `Referrer-Policy: no-referrer`;
- `Permissions-Policy`;
- HSTS en `EYEX_ENV=production`;
- HTTPS obligatorio en producción, considerando `X-Forwarded-Proto: https` cuando existe un proxy TLS;
- recuperación de panics sin stack trace en la respuesta;
- logs JSON mediante `slog`;
- límite público de requests con bypass opcional por API key;
- timeout de request;
- métricas Prometheus en `/metrics`.

Métricas expuestas:

```text
eyex_http_requests_total
eyex_http_errors_total
eyex_http_error_rate
eyex_theme_requests_total{type="..."}
eyex_request_latency_p95_seconds
```

## Alcance de accesibilidad

EyeX cubre adaptación y simulación de **color** y ofrece una señal limitada de contraste de texto. No convierte por sí solo una aplicación en accesible.

Fuera de alcance de EyeX están, entre otros:

- navegación completa por teclado;
- orden de foco y focus traps;
- atributos ARIA;
- roles y semántica HTML;
- nombres accesibles;
- compatibilidad con lectores de pantalla;
- subtítulos, transcripciones y alternativas para contenido no textual;
- tamaño y espaciado tipográfico;
- reflow y zoom;
- accesibilidad de documentos, canvas, SVG complejo o contenido de terceros.

Una integración que aspire a WCAG debe evaluar esos aspectos por separado.

## Extensión de navegador: permisos y límites

La extensión está en `extension/browser` y usa Manifest V3.

Permisos actuales:

- `storage`: guarda `eyexType`, `eyexMode` y `eyexEnabled` mediante `storage.sync`;
- `content_scripts.matches: <all_urls>`: permite que `content.js` inserte el estilo EyeX en páginas web compatibles.

La extensión no solicita `tabs`, `history`, `cookies`, `webRequest`, `downloads`, `clipboardRead`, `clipboardWrite` ni acceso a credenciales. El código actual no envía el contenido de la página a la API.

Limitaciones relevantes:

- `<all_urls>` es un alcance amplio y debe justificarse al publicar la extensión;
- los navegadores bloquean content scripts en páginas internas como `chrome://`, `edge://` y otras superficies protegidas;
- algunas tiendas, páginas privilegiadas, visores PDF e iframes pueden imponer restricciones adicionales;
- estilos dentro de closed Shadow DOM no pueden ser reescritos desde el content script;
- CSS con alta especificidad, canvas, imágenes y contenido renderizado por WebGL no necesariamente cambian con la inyección CSS;
- la extensión aplica paletas locales y no ejecuta la nueva simulación pixel a pixel sobre la página.

Consulta también `extension/browser/README.md`.

## SDKs

Cada SDK tiene su propio README con instalación y ejemplo mínimo:

- `sdk/react/README.md`;
- `sdk/react-native/README.md`;
- `sdk/flutter/README.md`.

Los tres clientes incluyen temas, paletas personalizadas, sugerencia y los métodos `simulate` / `simulateBatch`.

También se mantiene `sdk-ts` como cliente TypeScript genérico del contrato `v1`.

## Ejecutar los cuatro backends

Las instrucciones completas están en `CONTRIBUTING.md`. Resumen:

```bash
# Go — :8080
EYEX_PORT=8080 go run ./cmd/api

# PHP — :8081
php -S 127.0.0.1:8081 implementations/php/public/index.php

# TypeScript/Node — :8082
cd implementations/typescript-node
npm install
npm run build
EYEX_PORT=8082 npm start

# Java — :8083
cd implementations/java
EYEX_PORT=8083 mvn spring-boot:run
```

## OpenAPI y compatibilidad

`openapi.yaml` es el contrato público versionado. `v1.2.0` agrega `/api/v1/simulate` y `/api/v1/simulate/batch` de forma aditiva; no modifica ni elimina las rutas de tema existentes.

El CI valida que las rutas históricas y las nuevas continúen presentes y ejecuta la paridad HTTP del contrato v1 —incluida la simulación— antes de considerar el workflow correcto.

## Versionado semántico y deprecación

EyeX usa Semantic Versioning `MAJOR.MINOR.PATCH`:

- `PATCH`: correcciones compatibles sin cambiar el contrato público;
- `MINOR`: capacidades aditivas compatibles, como los endpoints de simulación de `v1.2.0`;
- `MAJOR`: cambios incompatibles en rutas, campos requeridos, semántica o tipos.

Política para `/api/v1/`:

1. no se eliminan ni renombran rutas o campos existentes dentro de `v1` sin un periodo de deprecación;
2. una deprecación se anuncia en README y CHANGELOG y, cuando aplique, mediante headers `Deprecation`, `Sunset` y `Link`;
3. el periodo objetivo mínimo es 90 días antes de retirar comportamiento público;
4. una ruptura deliberada del contrato se publica bajo una nueva versión mayor de API, por ejemplo `/api/v2/`;
5. correcciones urgentes de seguridad pueden acortar el periodo si mantener el comportamiento representa un riesgo material;
6. añadir campos opcionales o nuevas rutas es compatible dentro de `v1`.

## Pruebas locales

```bash
go test ./...
go vet ./...
php -l implementations/php/public/index.php
php -l implementations/php/simulation.php
tsc --target ES2022 --module NodeNext --moduleResolution NodeNext --strict --noEmit implementations/typescript-node/src/simulation.ts
python3 -m py_compile tests/parity.py
```

Con las dependencias instaladas también deben ejecutarse los builds de TypeScript, Java, Vue y SDKs definidos en `.github/workflows/ci.yml`.

## Estructura principal

```text
cmd/api/                         backend Go
internal/                        handlers, configuración, middleware y temas
pkg/colorblind/                  simulación y utilidades cromáticas
implementations/php/             backend PHP
implementations/typescript-node/ backend Fastify
implementations/java/            backend Spring Boot
frontend/html-js/                playground principal
frontend/vue/                    cliente Vue
web/                             demo TypeScript ligera del contrato v1.2
sdk/react/                       SDK React
sdk/react-native/                SDK React Native
sdk/flutter/                     SDK Flutter
sdk-ts/                          SDK TypeScript genérico
extension/browser/               extensión Manifest V3
tests/parity.py                  paridad HTTP de los cuatro backends
openapi.yaml                     contrato OpenAPI 3.0.3
```

## Contribuir

Lee `CONTRIBUTING.md` antes de abrir cambios. Cualquier modificación de una ruta compartida debe actualizar, cuando corresponda, los cuatro backends, `openapi.yaml`, las pruebas de paridad, el README y el CHANGELOG.

## Licencia

MIT. Consulta `LICENSE`.
