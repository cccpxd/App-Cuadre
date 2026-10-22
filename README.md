# App Cuadre

Aplicación web para cuadrar caja en un corresponsal bancario.

## Tecnologías
- Go
- Gin
- PostgreSQL
- GORM
- Vercel

## Requisitos
- Go 1.22+
- PostgreSQL 12+

## Variables de entorno
- PORT: puerto de la aplicación. Por defecto `:8080`
- DATABASE_URL: cadena de conexión PostgreSQL

## Ejecutar localmente
```bash
go run .
```

## Ejecutar con Docker Compose
```bash
docker compose up --build
```

Esto levanta:
- PostgreSQL en el puerto 5432
- La app en el puerto 8080

Variables de entorno:
- `PORT=:8080`
- `DATABASE_URL=host=db user=postgres password=postgres dbname=appcuadre port=5432 sslmode=disable`

## Despliegue en Vercel
1. Conecta el repositorio.
2. Configura la variable `DATABASE_URL`.
3. Usa el archivo `vercel.json` ya incluido.

## Funcionalidades principales
- Cuadre de caja con 10 líneas base.
- Suma en tiempo real de valores y diferencia.
- Historial de cuadres por fecha.
- Módulo avícola con ingresos y egresos.
- Panel de configuración.
