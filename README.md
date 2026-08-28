<!-- Este README es la unica guia general del repositorio y explica como usar, integrar, probar y desplegar EyeX. -->
# EyeX

<img src="frontend/html-js/assets/logo.png" alt="Logo de EyeX" width="180">

EyeX es una API de temas de pantalla para interfaces que necesitan ofrecer opciones de color orientadas a daltonismo, baja visión y alto contraste. En vez de cambiar colores uno por uno, devuelve una paleta completa con fondo, superficie, texto, acciones y estados para aplicarla como tema global.

El proyecto incluye un backend principal en Go, implementaciones equivalentes en PHP, TypeScript con Node y Java, una web de prueba, un cliente Vue, un CSS estático, un widget embebible, una extensión de navegador y SDKs para React, React Native y Flutter.

EyeX no diagnostica condiciones visuales. El test rápido incluido es únicamente orientativo y no reemplaza una evaluación médica ni una auditoría completa de accesibilidad.

## Qué puede hacer

- Temas para `normal`, `protanopia`, `deuteranopia`, `tritanopia`, `achromatopsia` y `low_vision`.
- Intensidad `mild`, `moderate` o `severe`.
- Modo `dark` o `light`.
- Alto contraste combinable con cualquier tipo.
- Comprobación WCAG AA de `text` contra `background` y `surface` con umbral 4.5:1.
- Adaptación de una paleta de marca mediante `POST /api/v1/theme/custom`.
- Test orientativo de cuatro preguntas.
- CSS estático en `/eyex.css`.
- Widget embebible en `/eyex.js`.
- Contrato OpenAPI importable desde `/openapi.yaml` o desde el archivo `openapi.yaml` del repositorio.
- Errores JSON estandarizados en español e inglés.
- CORS explícito para consumo desde navegadores.
- ETag y `Cache-Control` para temas.
- Límite público de solicitudes y API key opcional para omitir ese límite.
- Headers de seguridad y HTTPS obligatorio cuando `EYEX_ENV=production`.
- Logs JSON por request y métricas Prometheus en `/metrics`.
- Pruebas automáticas de paridad entre Go, PHP, TypeScript/Node y Java.

## Ejecutar localmente con Docker Desktop

Desde la raíz del repositorio:

```bash
docker build -t eyex .
docker run --rm --name eyex -p 8080:8080 eyex
```

Abre:

```text
http://localhost:8080
```

La web y la API se sirven desde el mismo proceso Go.

También puedes ejecutar el backend principal sin Docker:

```bash
go run ./cmd/api
```

## Endpoints principales

### Obtener tipos disponibles

```http
GET /api/v1/theme/types
```

Respuesta:

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

```http
GET /api/v1/theme/{type}
```

El contrato original sigue funcionando sin parámetros. Ejemplo:

```http
GET /api/v1/theme/deuteranopia
```

```json
{
  "type": "deuteranopia",
  "palette": {
    "background": "#1E1E1E",
    "surface": "#2A2A2A",
    "text": "#F5F5F5",
    "primary": "#4A90D9",
    "secondary": "#D9A24A",
    "error": "#D94A4A",
    "success": "#4AD98C"
  },
  "contrast_ok": true
}
```

Parámetros opcionales:

| Parámetro | Valores | Uso |
| --- | --- | --- |
| `severity` | `mild`, `moderate`, `severe` | Intensidad de la adaptación. |
| `mode` | `dark`, `light` | Variante clara u oscura. |
| `high_contrast` | `true`, `false` | Refuerza el contraste del tema. |

Ejemplo:

```http
GET /api/v1/theme/protanopia?severity=moderate&mode=light&high_contrast=true
```

`low_vision` prioriza alto contraste de manera predeterminada.

### Adaptar una paleta propia

```http
POST /api/v1/theme/custom
Content-Type: application/json
```

```json
{
  "type": "deuteranopia",
  "severity": "moderate",
  "mode": "dark",
  "high_contrast": true,
  "palette": {
    "background": "#101820",
    "surface": "#182430",
    "text": "#F8F9FA",
    "primary": "#E63946",
    "secondary": "#2A9D8F",
    "error": "#D62828",
    "success": "#2A9D8F"
  }
}
```

Todos los colores de entrada deben usar `#RRGGBB`.

### Test orientativo

```http
POST /api/v1/test/suggest
Content-Type: application/json
```

```json
{
  "answers": {
    "reds_look_darker": false,
    "green_brown_confusion": false,
    "blue_yellow_confusion": true,
    "colors_look_gray": false
  }
}
```

El resultado sugiere qué tema probar primero y siempre deja claro que no es un diagnóstico médico.

## Errores estandarizados

Todo error controlado responde únicamente con:

```json
{
  "error": "invalid_type",
  "message": "Tipo de daltonismo no soportado"
}
```

Códigos públicos:

| Código | HTTP | Significado |
| --- | ---: | --- |
| `invalid_type` | 400 | El tipo no pertenece a los tipos soportados o la ruta del tipo es sospechosa. |
| `invalid_parameter` | 400 | Un query parameter no tiene un valor permitido. |
| `invalid_request` | 400 | El JSON está mal formado, es demasiado grande o no cumple la estructura esperada. |
| `invalid_palette` | 400 | La paleta personalizada es incompleta o contiene colores inválidos. |
| `not_found` | 404 | La ruta o recurso solicitado no existe. |
| `method_not_allowed` | 405 | La ruta existe pero no admite el método HTTP utilizado. |
| `https_required` | 426 | En producción la petición no llegó mediante HTTPS. |
| `rate_limited` | 429 | Se agotó el límite público por minuto. |
| `internal_server_error` | 500 | Error interno sin exponer stack trace. |
| `timeout` | 504 | La operación superó el tiempo máximo configurado. |

## Español e inglés

EyeX usa `Accept-Language` para el campo `message` de los errores.

```bash
curl -H "Accept-Language: es" http://localhost:8080/api/v1/theme/no-existe
curl -H "Accept-Language: en" http://localhost:8080/api/v1/theme/no-existe
```

En inglés el mismo error devuelve, por ejemplo:

```json
{
  "error": "invalid_type",
  "message": "Unsupported color vision type"
}
```

## Caché HTTP

`GET /api/v1/theme/{type}` devuelve `ETag` y `Cache-Control`.

```bash
curl -i http://localhost:8080/api/v1/theme/deuteranopia
```

Con el ETag recibido:

```bash
curl -i \
  -H 'If-None-Match: "ETAG_RECIBIDO"' \
  http://localhost:8080/api/v1/theme/deuteranopia
```

Si el contenido no cambió, EyeX responde `304 Not Modified` sin reenviar el JSON.

## CORS

El backend principal permite consumo desde navegador y responde preflight `OPTIONS`. Por defecto `EYEX_ALLOWED_ORIGIN=*`.

Headers permitidos:

```text
Accept
Accept-Language
Content-Type
If-None-Match
X-API-Key
```

## Límite de solicitudes y API key

Sin API key se aplica `EYEX_RATE_LIMIT_PER_MINUTE`. El valor local predeterminado es 60 solicitudes por minuto y dirección IP.

Cuando se supera el límite:

```json
{
  "error": "rate_limited",
  "message": "Límite de solicitudes excedido"
}
```

Para una integración autorizada, el operador puede configurar `EYEX_API_KEY` y el cliente envía:

```http
X-API-Key: TU_CLAVE
```

Con una clave válida no se aplica el límite público. No guardes una clave real en GitHub; configúrala como secreto del proveedor de hosting.

## Seguridad

EyeX agrega, entre otros:

```text
Content-Security-Policy
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
Referrer-Policy: no-referrer
Permissions-Policy
```

Cuando `EYEX_ENV=production`, también agrega:

```text
Strict-Transport-Security: max-age=31536000; includeSubDomains
```

En producción una petición que no llegue por TLS o por un proxy que indique `X-Forwarded-Proto: https` recibe `426 https_required`.

Las rutas de tipo rechazan traversal, barras codificadas, backslashes y segmentos `..` antes de consultar la lógica del tema.

## Métricas y logs

```http
GET /metrics
```

El formato es compatible con Prometheus e incluye al menos:

```text
eyex_http_requests_total
eyex_http_errors_total
eyex_http_error_rate
eyex_theme_requests_total{type="..."}
eyex_request_latency_p95_seconds
```

El servidor Go escribe un objeto JSON por request con método, ruta, status, latencia, tipo y dirección remota.

## OpenAPI, Postman y Swagger

El contrato completo está en:

```text
openapi.yaml
```

Con el servidor Go activo también está disponible en:

```text
http://localhost:8080/openapi.yaml
```

Puedes importarlo directamente en Postman o Swagger UI. El CI ejecuta Spectral sobre este archivo para impedir que un contrato inválido llegue a `main`.

## Playground web

La página principal consulta la API real del mismo host. No utiliza paletas mockeadas para el demo. Permite elegir tipo, intensidad, modo y alto contraste y aplica el resultado como variables CSS a la propia interfaz.

```text
http://localhost:8080/
```

## CSS estático

```html
<link rel="stylesheet" href="https://TU_HOST/eyex.css">
```

Luego:

```html
<html data-eyex-theme="deuteranopia" data-eyex-mode="dark">
```

Para alto contraste:

```html
<html data-eyex-theme="deuteranopia" data-eyex-mode="dark" data-eyex-high-contrast="true">
```

## Widget embebible

Cuando el widget y la API están publicados en el mismo host:

```html
<script src="https://TU_HOST/eyex.js"></script>
```

El script agrega el selector flotante y consulta la API de EyeX.

## Implementaciones incluidas

| Implementación | Ruta | Rol |
| --- | --- | --- |
| Go | `cmd/api` + `internal` | Backend oficial y servidor de la web. |
| PHP | `implementations/php` | Backend equivalente. |
| TypeScript/Node | `implementations/typescript-node` | Backend Fastify equivalente. |
| Java | `implementations/java` | Backend Spring Boot equivalente. |
| HTML + JS | `frontend/html-js` | Playground principal. |
| Vue | `frontend/vue` | Cliente alternativo. |
| React | `sdk/react` | SDK web. |
| React Native | `sdk/react-native` | SDK móvil JavaScript. |
| Flutter | `sdk/flutter` | SDK móvil Dart. |
| Extensión | `extension/browser` | Aplicación de temas desde el navegador. |

## Comprobar paridad entre backends

El archivo `tests/parity.py` llama a las cuatro implementaciones con los mismos datos y falla si el status o el JSON no son idénticos.

El CI levanta Go, PHP, TypeScript/Node y Java en puertos independientes y ejecuta automáticamente esa comparación.

Para validar el backend oficial por HTTP:

```bash
./tests/smoke.sh http://localhost:8080
```

## Configuración

El proyecto mantiene un único archivo `.env` local. `.gitignore` evita que se publique accidentalmente en GitHub.

Variables reconocidas:

| Variable | Predeterminado | Función |
| --- | --- | --- |
| `EYEX_PORT` | `8080` | Puerto local. Si no está definido, Go también reconoce `PORT` del hosting. |
| `EYEX_ALLOWED_ORIGIN` | `*` | Origin permitido por CORS. |
| `EYEX_ENV` | `development` | Usa `production` para habilitar HTTPS obligatorio y HSTS. |
| `EYEX_RATE_LIMIT_PER_MINUTE` | `60` | Límite público por IP. |
| `EYEX_REQUEST_TIMEOUT_MS` | `5000` | Tiempo máximo de ejecución del request. |
| `EYEX_API_KEY` | vacío | Clave opcional que omite el rate limit. |

## Publicar desde GitHub con Render

El repositorio incluye `render.yaml` y `Dockerfile`.

Después de subir EyeX a GitHub:

1. En Render selecciona **New > Blueprint**.
2. Conecta el repositorio de EyeX.
3. Render detectará `render.yaml`.
4. Crea el servicio.
5. Si quieres integraciones sin límite público, configura `EYEX_API_KEY` como secreto desde el panel de Render.

La misma URL pública servirá la web y la API:

```text
https://TU_HOST/
https://TU_HOST/api/v1/theme/types
https://TU_HOST/api/v1/theme/deuteranopia
https://TU_HOST/openapi.yaml
https://TU_HOST/metrics
```

Los manifiestos Kubernetes permanecen en `deploy/k8s` y Terraform en `deploy/terraform` para despliegues administrados por infraestructura.

## Integración desde JavaScript

```javascript
const response = await fetch(
  'https://TU_HOST/api/v1/theme/deuteranopia?severity=moderate&mode=dark'
);
const data = await response.json();

for (const [name, value] of Object.entries(data.palette)) {
  document.documentElement.style.setProperty(`--eyex-${name}`, value);
}
```

Luego:

```css
body {
  background: var(--eyex-background);
  color: var(--eyex-text);
}

.card {
  background: var(--eyex-surface);
}
```

## Contribuir y licencia

Las instrucciones de colaboración están en `CONTRIBUTING.md`. El repositorio utiliza licencia MIT, disponible en `LICENSE`.

Al subir el proyecto a GitHub, el workflow de gobernanza crea de forma idempotente la etiqueta `good first issue` y tres tareas iniciales para nuevos colaboradores.
