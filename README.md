# EyeX

<img src="frontend/html-js/assets/logo.png" alt="Logo de EyeX" width="180">

EyeX es una API de temas de pantalla pensada para interfaces que necesitan ofrecer opciones de color para personas con daltonismo o baja visión. En lugar de modificar colores uno por uno, devuelve una paleta completa con fondo, superficie, texto, acciones y estados para aplicarla a toda la interfaz.

El proyecto incluye un backend principal en Go, implementaciones equivalentes en PHP, TypeScript con Node y Java, una web de demostración, un cliente Vue, un CSS estático, un widget embebible, una extensión de navegador y SDKs para React, React Native y Flutter.

EyeX no diagnostica condiciones visuales. El test incluido es únicamente orientativo y no reemplaza una evaluación médica ni una auditoría completa de accesibilidad.

## Funciones principales

- Temas para `normal`, `protanopia`, `deuteranopia`, `tritanopia`, `achromatopsia` y `low_vision`.
- Intensidad configurable con `mild`, `moderate` y `severe`.
- Variante `dark` o `light`.
- Opción de alto contraste combinable con cualquier tipo de visión.
- Validación de contraste WCAG AA para texto sobre `background` y `surface` con umbral mínimo de 4.5:1.
- Adaptación de una paleta de marca enviada por el cliente.
- Test orientativo de cuatro preguntas.
- CSS estático para usar EyeX sin consumir la API.
- Widget flotante integrable con una sola etiqueta `<script>`.
- Extensión compatible con navegadores basados en Chromium y preparada para Firefox.
- SDKs para React, React Native y Flutter.
- Misma lógica de API disponible en Go, PHP, TypeScript/Node y Java.

## Ejecutar EyeX con Docker Desktop

Desde la raíz del repositorio:

```bash
docker build -t eyex:local .
docker run --rm --name eyex -p 8080:8080 eyex:local
```

Abre:

```text
http://localhost:8080
```

La web principal y la API se sirven desde el mismo proceso Go.

## Contrato base compatible

### Obtener un tema

```http
GET /api/v1/theme/{type}
```

Los clientes que ya consumían esta ruta pueden seguir haciéndolo sin parámetros. Las paletas originales para los cinco tipos iniciales se mantienen sin cambios. La respuesta incorpora el campo aditivo `contrast_ok`.

Ejemplo:

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

Si el tipo no existe:

```json
{
  "error": "invalid_type",
  "message": "Tipo de daltonismo no soportado"
}
```

### Consultar tipos disponibles

```http
GET /api/v1/theme/types
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

## Intensidad, tema claro/oscuro y alto contraste

`GET /api/v1/theme/{type}` acepta parámetros opcionales:

| Parámetro | Valores | Uso |
| --- | --- | --- |
| `severity` | `mild`, `moderate`, `severe` | Define cuánto se aleja la paleta de la apariencia normal hacia la paleta optimizada. |
| `mode` | `dark`, `light` | Selecciona un tema oscuro o claro. |
| `high_contrast` | `true`, `false` | Refuerza contraste y puede combinarse con cualquier tipo. |

Ejemplo:

```http
GET /api/v1/theme/protanopia?severity=moderate&mode=light&high_contrast=true
```

Cuando no se envían estos parámetros, EyeX conserva el comportamiento y los colores originales del contrato base.

`low_vision` prioriza alto contraste de manera predeterminada.

## Qué significa `contrast_ok`

EyeX calcula la luminancia relativa y la relación de contraste siguiendo la fórmula de WCAG. `contrast_ok` es `true` cuando el color `text` alcanza al menos 4.5:1 tanto contra `background` como contra `surface`.

Este indicador no significa que toda una aplicación cumpla WCAG. Imágenes, tamaños de fuente, iconos, estados de foco, estructura semántica y otros componentes siguen requiriendo revisión propia.

## Adaptar una paleta de marca

```http
POST /api/v1/theme/custom
Content-Type: application/json
```

Ejemplo:

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

EyeX conserva los colores base cuando es razonable, adapta los colores semánticos al tipo elegido, ajusta el modo claro u oscuro y corrige el color de texto cuando sea necesario para alcanzar el umbral configurado.

La respuesta usa el mismo formato que `GET /api/v1/theme/{type}`.

Todos los colores de entrada deben usar formato `#RRGGBB`.

## Test orientativo

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

Respuesta:

```json
{
  "suggested_type": "tritanopia",
  "disclaimer": "Resultado orientativo. No es un diagnóstico médico."
}
```

Las cuatro preguntas sirven únicamente para sugerir qué tema probar primero. No realizan un diagnóstico.

## Aplicar una paleta con JavaScript

```javascript
const response = await fetch(
  '/api/v1/theme/deuteranopia?severity=moderate&mode=dark'
);
const data = await response.json();

for (const [name, value] of Object.entries(data.palette)) {
  document.documentElement.style.setProperty(`--eyex-${name}`, value);
}
```

Después los componentes pueden usar las variables:

```css
body {
  background: var(--eyex-background);
  color: var(--eyex-text);
}

.card {
  background: var(--eyex-surface);
}

.button-primary {
  background: var(--eyex-primary);
}
```

## CSS estático

EyeX sirve un archivo preparado en:

```text
/eyex.css
```

Integración:

```html
<link rel="stylesheet" href="https://TU_HOST/eyex.css">
```

Luego selecciona el tema en el elemento raíz:

```html
<html data-eyex-theme="deuteranopia" data-eyex-mode="dark">
```

Para alto contraste:

```html
<html
  data-eyex-theme="deuteranopia"
  data-eyex-mode="dark"
  data-eyex-high-contrast="true">
```

El CSS estático usa paletas predefinidas y no realiza solicitudes HTTP.

## Widget embebible

EyeX también sirve:

```text
/eyex.js
```

Cuando el widget y la API están publicados en el mismo host, basta una línea:

```html
<script src="https://TU_HOST/eyex.js"></script>
```

El script agrega un botón flotante con selector de tipo, intensidad, tema y alto contraste. Por defecto toma como API el mismo origen desde el que se descargó `eyex.js`.

Si la API está en otro host:

```html
<script
  src="https://TU_CDN/eyex.js"
  data-eyex-api="https://api.ejemplo.com"></script>
```

También expone:

```javascript
await window.EyeX.apply('protanopia', {
  severity: 'moderate',
  mode: 'dark',
  highContrast: true
});
```

## Extensión de navegador

La extensión se encuentra en:

```text
extension/browser
```

Para probarla en Chrome o Edge:

1. Abre la página de extensiones del navegador.
2. Activa el modo de desarrollador.
3. Selecciona **Cargar descomprimida**.
4. Selecciona la carpeta `extension/browser`.
5. Abre EyeX desde el icono de extensiones y activa el modo deseado.

El mismo código evita dependencias del backend y guarda la preferencia mediante `storage.sync`.

Para probarla temporalmente en Firefox, abre `about:debugging`, entra en **Este Firefox**, selecciona **Cargar complemento temporal** y elige `extension/browser/manifest.json`.

## SDK para React

Ubicación:

```text
sdk/react
```

Ejemplo:

```tsx
import { EyeXClient, useEyeXTheme } from '@eyex/react';

const client = new EyeXClient('https://api.ejemplo.com');

function AccessibilityTheme() {
  const { type, setType, theme } = useEyeXTheme(client, 'normal', {
    severity: 'moderate',
    mode: 'dark'
  });

  return (
    <select value={type} onChange={e => setType(e.target.value as typeof type)}>
      <option value="normal">Normal</option>
      <option value="deuteranopia">Deuteranopia</option>
    </select>
  );
}
```

El cliente React incluye `types()`, `theme()`, `custom()` y `suggest()`. El hook aplica automáticamente la paleta como variables CSS globales.

## SDK para React Native

Ubicación:

```text
sdk/react-native
```

El cliente incluye `types()`, `theme()`, `custom()` y `suggest()`, y devuelve la paleta para aplicarla mediante `StyleSheet` o el sistema de temas de la aplicación:

```typescript
const client = new EyeXClient('https://api.ejemplo.com');
const theme = await client.theme('tritanopia', {
  severity: 'moderate',
  mode: 'dark'
});
```

## SDK para Flutter

Ubicación:

```text
sdk/flutter
```

Ejemplo:

```dart
final client = EyeXClient('https://api.ejemplo.com');
final theme = await client.theme(
  'protanopia',
  severity: 'moderate',
  mode: 'dark',
);
```

El cliente Flutter incluye `types()`, `theme()`, `custom()` y `suggest()`. La aplicación puede convertir los valores `#RRGGBB` recibidos a objetos `Color` y construir su `ThemeData`.

## Implementaciones backend

| Implementación | Ubicación | Tecnología |
| --- | --- | --- |
| Go | raíz del proyecto | `net/http` |
| PHP | `implementations/php` | PHP 8 sin framework |
| TypeScript | `implementations/typescript-node` | Node.js + Fastify |
| Java | `implementations/java` | Spring Boot |

Las cuatro implementaciones mantienen los mismos tipos, parámetros, endpoints, reglas de contraste y estructura JSON.

## Clientes web

| Cliente | Ubicación | Uso |
| --- | --- | --- |
| HTML + JavaScript | `frontend/html-js` | Web principal servida por Go |
| Vue | `frontend/vue` | Cliente alternativo con Vue y composable |

## Ejecutar Go sin Docker

Requisito: Go 1.23 o superior.

```bash
go test ./...
go run ./cmd/api
```

## Ejecutar PHP

```bash
cd implementations/php
php -S 127.0.0.1:8080 public/index.php
```

## Ejecutar TypeScript / Node

```bash
cd implementations/typescript-node
npm install
npm run build
npm start
```

## Ejecutar Java

```bash
cd implementations/java
mvn spring-boot:run
```

## Ejecutar Vue

Primero deja un backend escuchando en `http://localhost:8080`.

```bash
cd frontend/vue
npm install
npm run dev
```

Si la API está en otra dirección:

```text
http://localhost:5173/?api=https://api.ejemplo.com
```

No necesita otro `.env`.

## Configuración

El repositorio mantiene un único archivo `.env` en la raíz:

```dotenv
EYEX_PORT=8080
EYEX_ALLOWED_ORIGIN=*
```

No se requieren archivos `.env` adicionales para los SDK, Vue, PHP, Java o Node.

`EYEX_PORT` define el puerto del servicio Go/Node/PHP cuando corresponde. `EYEX_ALLOWED_ORIGIN` controla CORS. Para una publicación real conviene sustituir `*` por el origen que deba consumir la API cuando frontend y API estén separados.

## Estructura

```text
eyex/
├── cmd/api/                       # Arranque del backend Go
├── internal/                      # Temas, contraste, API y middleware Go
├── implementations/
│   ├── php/
│   ├── typescript-node/
│   └── java/
├── frontend/
│   ├── html-js/                   # Web, eyex.css y eyex.js
│   └── vue/
├── extension/browser/             # Extensión Chrome/Firefox
├── sdk/
│   ├── react/
│   ├── react-native/
│   └── flutter/
├── deploy/                        # Kubernetes y Terraform de referencia
├── .env                           # Único archivo de variables
├── Dockerfile
├── go.mod
└── README.md                      # Único archivo Markdown
```

## Guardar el proyecto en GitHub

Crea un repositorio vacío en GitHub y, desde la carpeta `eyex`, ejecuta:

```bash
git init
git add .
git commit -m "Initial EyeX"
git branch -M main
git remote add origin https://github.com/TU_USUARIO/eyex.git
git push -u origin main
```

Si el repositorio ya existe localmente, no repitas `git init`; agrega o corrige el `origin` y realiza el `push`.

El `Dockerfile` de la raíz es suficiente para conectar el repositorio a un hosting compatible con Docker. La imagen oficial ejecuta el backend Go y sirve también la web principal, `eyex.css`, `eyex.js` y el logo desde el mismo host.

## Alcance de accesibilidad

EyeX ayuda a aplicar un tema visual consistente y valida el contraste principal de texto. No modifica imágenes ni contenido multimedia, no identifica automáticamente una condición visual y no garantiza por sí solo el cumplimiento integral de WCAG. Una aplicación accesible también debe considerar estructura semántica, navegación por teclado, foco, tamaños de texto, estados, formularios, contenido alternativo y pruebas con usuarios.
