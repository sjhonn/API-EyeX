# @eyex/react

Cliente React para EyeX API `v1.2.0`.

## Instalación

Desde el repositorio:

```bash
cd sdk/react
npm install
npm run build
```

Cuando el paquete se publique en un registro:

```bash
npm install @eyex/react
```

## Uso mínimo

```tsx
import { EyeXClient, useEyeXTheme } from '@eyex/react';

const client = new EyeXClient('http://localhost:8080');

export function AccessibilityTheme() {
  const { type, setType, theme, loading } = useEyeXTheme(client, 'normal', {
    severity: 'moderate',
    mode: 'dark',
  });

  if (loading) return <span>Cargando…</span>;

  return (
    <select value={type} onChange={(event) => setType(event.target.value as typeof type)}>
      <option value="normal">Normal</option>
      <option value="deuteranopia">Deuteranopia</option>
    </select>
  );
}
```

El hook aplica la paleta recibida como variables CSS `--eyex-*`.

## Simular un color

```ts
const result = await client.simulate('#FF0000', 'protanopia', 0.65);
console.log(result.simulated); // #A05A00
```

## Simular un lote

```ts
const result = await client.simulateBatch(
  ['#FF0000', '#00FF00', '#0000FF'],
  'deuteranopia',
  0.5,
);
```

Métodos disponibles: `types`, `theme`, `custom`, `suggest`, `simulate` y `simulateBatch`.
