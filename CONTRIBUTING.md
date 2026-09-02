# Contribuir a EyeX

Este repositorio mantiene un contrato HTTP compartido entre Go, PHP, TypeScript/Node y Java. Una modificación de API no se considera terminada hasta que el contrato, las implementaciones afectadas, la documentación y la paridad estén alineados.

## Requisitos

Herramientas recomendadas:

- Go 1.23 o superior;
- PHP 8.4 o compatible con las características usadas por el proyecto;
- Node.js 22 y npm;
- Java 21 y Maven 3.9+;
- Dart 3.4+ para el SDK Flutter;
- Python 3.11+ para `tests/parity.py`.

El proyecto utiliza un único `.env` en la raíz. No crees archivos `.env` por backend, frontend o SDK. Todos los runtimes que leen configuración local deben usar el archivo raíz o variables de entorno del proceso.

## Preparación

Desde la raíz:

```bash
go test ./...
go vet ./...
```

Para dependencias Node:

```bash
cd implementations/typescript-node && npm install && cd ../..
cd frontend/vue && npm install && cd ../..
cd web && npm install && cd ..
cd sdk/react && npm install && cd ../..
cd sdk/react-native && npm install && cd ../..
cd sdk-ts && npm install && cd ..
```

Para Java:

```bash
cd implementations/java
mvn test
cd ../..
```

Para Flutter:

```bash
cd sdk/flutter
dart pub get
dart analyze
cd ../..
```

## Levantar los cuatro backends localmente

Usa puertos distintos para poder ejecutar la paridad.

### Go — puerto 18080

Desde la raíz:

```bash
EYEX_PORT=18080 go run ./cmd/api
```

### PHP — puerto 18081

Desde la raíz:

```bash
php -S 127.0.0.1:18081 implementations/php/public/index.php
```

El router PHP carga el `.env` de la raíz; el puerto lo controla el servidor embebido de PHP.

### TypeScript/Node — puerto 18082

```bash
cd implementations/typescript-node
npm install
npm run build
EYEX_PORT=18082 npm start
```

Para desarrollo interactivo:

```bash
EYEX_PORT=18082 npm run dev
```

### Java — puerto 18083

```bash
cd implementations/java
EYEX_PORT=18083 mvn spring-boot:run
```

`application.properties` toma el puerto desde `EYEX_PORT`.

## Ejecutar la paridad

Con los cuatro procesos levantados:

```bash
python3 tests/parity.py \
  --go http://127.0.0.1:18080 \
  --php http://127.0.0.1:18081 \
  --typescript http://127.0.0.1:18082 \
  --java http://127.0.0.1:18083
```

La prueba compara status y JSON. Para simulación también valida vectores conocidos de salida.

## Reglas para cambios de API

Antes de modificar el contrato:

1. revisa `openapi.yaml` y conserva rutas/campos existentes salvo que exista una deprecación formal;
2. implementa el cambio en Go, PHP, TypeScript/Node y Java cuando sea una capacidad compartida;
3. agrega pruebas unitarias donde exista infraestructura de tests;
4. actualiza `tests/parity.py` con al menos un caso positivo y errores relevantes;
5. actualiza los SDKs si el endpoint debe estar disponible para clientes;
6. actualiza README y CHANGELOG;
7. ejecuta las validaciones locales disponibles;
8. confirma que `EyeX CI` finaliza correctamente.

Cambios aditivos compatibles pertenecen a una versión `MINOR`. Cambios incompatibles requieren una nueva versión mayor del API y la política de deprecación descrita en el README.

## Simulación de color

Las matrices de Machado deben mantenerse idénticas en las cuatro implementaciones. No modifiques coeficientes, orden de canales, gamma, interpolación o redondeo en un solo backend.

La secuencia canónica es:

1. `#RRGGBB` a sRGB normalizado;
2. sRGB a RGB lineal;
3. interpolación de matriz de severidad;
4. multiplicación matricial;
5. clamp a `[0,1]`;
6. RGB lineal a sRGB;
7. redondeo a 8 bits y hexadecimal en mayúsculas.

`severity` admite `[0,1]`; si se omite, vale `1`.

## OpenAPI

`openapi.yaml` debe continuar siendo YAML válido y contener al menos todas las rutas históricas de `v1` además de cualquier ruta nueva. El CI ejecuta una validación de contrato y mantiene la paridad HTTP.

## Estilo y validación

Go:

```bash
gofmt -w $(find cmd internal pkg -name '*.go' -type f)
go test ./...
go vet ./...
```

PHP:

```bash
php -l implementations/php/public/index.php
php -l implementations/php/simulation.php
```

TypeScript/Node:

```bash
cd implementations/typescript-node
npm run build
```

Java:

```bash
cd implementations/java
mvn test
```

SDK TypeScript genérico:

```bash
npm --prefix sdk-ts run build
```

No agregues artefactos de build, `node_modules`, `target`, caches ni archivos `.env` adicionales al commit.
