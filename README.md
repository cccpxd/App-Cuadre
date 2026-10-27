# App Cuadre

Aplicación web para cuadrar caja en un corresponsal bancario, con módulos de
avícola, ganancias avícolas y consulta de ganancias por periodos.

## Tecnologías
- Go
- Gin
- PostgreSQL
- GORM
- Vercel

## Requisitos
- Go 1.25+
- PostgreSQL 12+

## Variables de entorno
- `PORT`: puerto de la aplicación. Por defecto `:8080`.
- `DATABASE_URL`: cadena de conexión PostgreSQL.
- `APP_TIMEZONE`: zona horaria de respaldo para consultas. Por defecto
  `America/Bogota`.
- `RUN_MIGRATIONS`: si es `false`, no crea ni actualiza las tablas al arrancar.
  Por defecto `true`. Úsalo en Vercel cuando las tablas ya existan.
- `ADMIN_EMAIL` / `ADMIN_PASSWORD`: primer usuario.
- `ADMIN_EMAIL_2` / `ADMIN_PASSWORD_2`: segundo usuario opcional.

Los usuarios se crean automáticamente a partir de esas variables. Si un email
ya existe, su contraseña no se modifica. Los datos históricos que no tengan
dueño se asignan al primer usuario configurado (`ADMIN_EMAIL`); el segundo
usuario empieza sin datos. Cada usuario ve únicamente sus cuadres, movimientos,
configuración y ganancias.

## Ejecutar localmente
```bash
go run .
```

## Ejecutar con Docker Compose
```bash
docker compose up --build
```

Esto levanta:
- PostgreSQL en el puerto 5432.
- La app en el puerto 8080.

Variables de entorno de Docker Compose:
- `PORT=:8080`
- `DATABASE_URL=host=db user=postgres password=postgres dbname=appcuadre port=5432 sslmode=disable`
- `APP_TIMEZONE=America/Bogota`
- `ADMIN_EMAIL`, `ADMIN_PASSWORD`, `ADMIN_EMAIL_2`, `ADMIN_PASSWORD_2`

## Despliegue en Vercel
1. Conecta el repositorio.
2. Configura `DATABASE_URL`, `APP_TIMEZONE` y los usuarios en las variables de
   entorno del proyecto.
3. Usa el archivo `vercel.json` ya incluido.

## Funcionalidades principales
- Cuadre de caja con 10 líneas base.
- Suma en tiempo real de valores y diferencia.
- Historial de cuadres por fecha.
- Módulo avícola con ingresos y egresos.
- Ganancias Avícola independiente del módulo avícola de saldos:
  - Ganancia: se calcula sola como caja principal + caja auxiliar.
  - Cuadres de caja principal.
  - Cuadres de caja auxiliar.
  - Campos adicionales por sección con totales en vivo.
- Consulta de Ganancia Avícola:
  - Semana, quincena y mes actual.
  - Últimos 7, 15 y 30 días.
  - Rango manual de fechas.
  - Totales y detalle de registros.
- Aislamiento de datos por usuario autenticado.
- Panel de configuración por usuario.

## Pruebas
```bash
go test ./...
```
