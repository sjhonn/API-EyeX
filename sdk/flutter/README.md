# EyeX Flutter SDK

Cliente Dart para EyeX API `v1.2.0`.

## Instalación

Para trabajar dentro del repositorio:

```bash
cd sdk/flutter
dart pub get
dart analyze
```

Para consumirlo como dependencia local desde otra aplicación:

```yaml
dependencies:
  eyex:
    path: ../API-EyeX/sdk/flutter
```

## Uso mínimo

```dart
import 'package:eyex/eyex.dart';

final client = EyeXClient('http://localhost:8080');

final theme = await client.theme(
  'deuteranopia',
  severity: 'moderate',
  mode: 'dark',
);

print(theme.palette.primary);
```

## Simulación

```dart
final simulated = await client.simulate(
  '#FF0000',
  'protanopia',
  severity: 0.65,
);

print(simulated.simulated); // #A05A00
```

Batch:

```dart
final batch = await client.simulateBatch(
  ['#FF0000', '#00FF00', '#0000FF'],
  'deuteranopia',
  severity: 0.5,
);
```

Métodos disponibles: `types`, `theme`, `custom`, `suggest`, `simulate` y `simulateBatch`.
