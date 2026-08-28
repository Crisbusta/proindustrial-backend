# ProIndustrial Backend Public

API pública y backoffice de aprobación para ProIndustrial.

## Qué incluye

- API pública para categorías, empresas, cotizaciones y registros
- Login de proveedores y panel protegido por JWT
- Backoffice admin para aprobar o rechazar registros de nuevas empresas
- Alta automática de `company` + `user` al aprobar
- Cambio obligatorio de contraseña en el primer login del proveedor
- Envío opcional de correo de activación por Resend API o SMTP

## Requisitos

- Go 1.25+
- Postgres

## Variables de entorno

- `DATABASE_URL`
- `JWT_SECRET`
- `PORT`
- `CORS_ORIGIN`
- `APP_BASE_URL`
- `RESEND_API_KEY`
- `RESEND_FROM`
- `SMTP_HOST`
- `SMTP_PORT`
- `SMTP_USER`
- `SMTP_PASS`
- `SMTP_FROM`

Ejemplo base en [.env.example](/Users/coto/Desktop/marketplace-b2b/backend-public/.env.example).

## Ejecutar en local

Con Docker desde este directorio:

```bash
docker compose up --build
```

O con Go:

```bash
go test ./...
go run ./cmd/server
```

## Endpoints principales

Públicos:

- `GET /api/category-groups`
- `GET /api/regions`
- `GET /api/companies`
- `GET /api/companies/:slug`
- `POST /api/quotes`
- `POST /api/registrations`

Proveedor:

- `POST /api/auth/login`
- `GET /api/auth/me`
- `POST /api/auth/change-password`
- `GET /api/panel/dashboard/stats`
- `GET /api/panel/quotes`
- `GET/POST/PATCH/DELETE /api/panel/services`
- `GET /api/panel/profile`
- `PUT /api/panel/profile`

Admin:

- `POST /api/admin/auth/login`
- `GET /api/admin/auth/me`
- `GET /api/admin/registrations`
- `GET /api/admin/registrations/:id`
- `POST /api/admin/registrations/:id/approve`
- `POST /api/admin/registrations/:id/reject`

## Credenciales seed

Las cuentas sembradas por las migraciones son **solo para entorno local**.
Su contraseña no se versiona en el repositorio: pídela al responsable del
entorno o rótala con un hash bcrypt propio.

En producción, el usuario administrador debe tener una contraseña rotada
manualmente; la sembrada por la migración `003` nunca debe seguir en uso.

## Flujo de aprobación

1. Un proveedor envía `POST /api/registrations`.
2. El registro queda en `pending`.
3. Un admin entra al backoffice y aprueba.
4. El backend crea la empresa y el usuario proveedor.
5. El proveedor recibe una contraseña temporal generada al azar con `crypto/rand`.
6. En el primer login debe cambiarla antes de usar el panel.
7. Si `RESEND_API_KEY` y `RESEND_FROM` están configurados, el correo sale por Resend.
8. Si Resend no está configurado pero SMTP sí, usa SMTP.
9. Si no hay proveedor de correo configurado, el mensaje queda en logs.
