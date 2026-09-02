# EyeX Browser Extension

Extensión Manifest V3 que aplica una paleta EyeX mediante CSS inyectado y guarda la preferencia del usuario.

## Permisos

`manifest.json` declara:

```json
{
  "permissions": ["storage"],
  "content_scripts": [{ "matches": ["<all_urls>"] }]
}
```

`storage` se usa exclusivamente para `eyexType`, `eyexMode` y `eyexEnabled` mediante `storage.sync`.

`<all_urls>` permite ejecutar `content.js` en páginas web compatibles. Es un permiso de alcance amplio: debe justificarse de forma explícita ante la tienda del navegador.

La versión actual no solicita `tabs`, `history`, `cookies`, `webRequest`, `downloads`, permisos de portapapeles ni credenciales. Tampoco envía contenido de la página al backend EyeX.

## Límites

La extensión no puede garantizar modificaciones en:

- páginas internas/protegidas del navegador (`chrome://`, `edge://` y equivalentes);
- tiendas de extensiones y otras superficies privilegiadas;
- algunos visores PDF e iframes restringidos;
- closed Shadow DOM;
- canvas, imágenes, video o WebGL;
- componentes cuyo CSS no pueda ser sobrescrito razonablemente por las reglas inyectadas.

La extensión aplica paletas de interfaz. No procesa los pixeles de la página con `/api/v1/simulate` y no sustituye una auditoría de accesibilidad por teclado, ARIA o lector de pantalla.

## Carga local

Chrome/Edge:

1. abre la página de extensiones;
2. habilita modo desarrollador;
3. selecciona **Cargar descomprimida**;
4. selecciona `extension/browser`.

Firefox:

1. abre `about:debugging`;
2. entra en **Este Firefox**;
3. usa **Cargar complemento temporal**;
4. selecciona `manifest.json`.
