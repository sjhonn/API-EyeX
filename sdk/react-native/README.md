# @eyex/react-native

Cliente TypeScript para consumir EyeX API desde React Native.

## Instalación

Desde el repositorio:

```bash
cd sdk/react-native
npm install
npm run build
```

Cuando el paquete esté publicado:

```bash
npm install @eyex/react-native
```

El runtime debe proporcionar `fetch`, como ocurre en las versiones actuales de React Native.

## Uso mínimo

```ts
import { EyeXClient } from '@eyex/react-native';

const client = new EyeXClient('http://localhost:8080');

const theme = await client.theme('tritanopia', {
  severity: 'moderate',
  mode: 'dark',
});

console.log(theme.palette.primary);
```

## Simulación

```ts
const simulated = await client.simulate('#FF0000', 'protanopia', 0.65);
console.log(simulated.simulated); // #A05A00

const batch = await client.simulateBatch(
  ['#FF0000', '#00FF00'],
  'deuteranopia',
  0.5,
);
```

Métodos disponibles: `types`, `theme`, `custom`, `suggest`, `simulate` y `simulateBatch`.
