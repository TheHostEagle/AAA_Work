# AutoMarket · Taller AAA

API REST en **Go + Gin** para la venta de vehículos usados, con el marco de seguridad **AAA** (Authentication, Authorization, Accounting) sobre todas sus operaciones.

Todo corre en la máquina local: **sin base de datos, sin Docker, sin servidores externos**. La persistencia son archivos JSON en la carpeta `data/`.

## Integrantes:

1. Duque – Santiago
2. Escudero – Alejandro
3. Gomez – Dana
4. Herrera – Hilary

## Qué cumple el proyecto

| Requisito | Dónde se implementa |
|---|---|
| **HU-VIS-01** Consulta pública del catálogo con el contacto del vendedor | `GET /api/vehicles` (sin token) |
| **HU-VEN-01** Registro, publicación (queda `pendiente`) y reporte de venta | `POST /api/auth/register`, `POST /api/vehicles`, `PUT /api/vehicles/:id/sold` |
| **HU-ADM-01** Aprobar publicaciones y eliminar publicaciones o usuarios | `PUT /api/admin/vehicles/:id/approve`, `DELETE /api/admin/vehicles/:id`, `DELETE /api/admin/users/:id` |
| **RF-01 Authentication** | `AuthenticationMiddleware`: exige `Authorization: Bearer <token>`; responde 401 |
| **RF-01 Authorization** | `AuthorizationMiddleware`: valida el rol en cada petición; responde 403 |
| **RF-01 Accounting** | `AccountingMiddleware`: registra toda petición en `data/audit.json` con hash encadenado |

Restricciones del taller: solo la librería estándar de Go y **Gin como única dependencia externa**, HTTP sin TLS, sin JWT, bcrypt, ORMs ni drivers de base de datos.

## 1. Requisitos previos

| Requisito | Detalle |
|---|---|
| Go | 1.21 o superior (`go version`). Descarga: https://go.dev/dl/ |
| Git | Para clonar y cambiar de rama |
| Terminal | PowerShell (los comandos están escritos para PowerShell) |
| Internet | Solo la primera vez, para que Go descargue Gin y sus dependencias |
| Puerto | El servidor usa el **8080** (fijo en `cmd/server/main.go`); debe estar libre |

## 2. Obtener el código

```powershell
git clone https://github.com/TheHostEagle/AAA_Work.git
cd AAA_Work
git checkout develop
```

Todos los comandos siguientes se ejecutan desde la raíz del repositorio (donde está `go.mod`).

## 3. Administrador inicial (variables de entorno)

El registro público solo crea **vendedores**. El único administrador se crea al arrancar el servidor a partir de dos variables de entorno. Si no se definen, el servidor arranca igual, pero las rutas `/api/admin/...` no se pueden usar.

```powershell
$env:ADMIN_EMAIL = "admin@automarket.test"
$env:ADMIN_PASSWORD = "Admin-demo-2026"
```

- Son valores de ejemplo para la máquina local; **no los subas al repositorio**.
- Valen solo para esa ventana de PowerShell: defínelas en la misma ventana donde ejecutas el servidor.
- La cuenta se crea una sola vez; en los siguientes arranques se reutiliza.

## 4. Ejecutar

```powershell
go mod tidy
go vet ./...
go test ./...
go run ./cmd/server
```

- `go mod tidy` no debe modificar `go.mod` ni `go.sum` (`git diff go.mod go.sum` sale vacío).
- `go test ./...` debe mostrar `ok` en cada paquete.
- `go run ./cmd/server` deja el servidor corriendo en esa ventana. Para detenerlo: `Ctrl+C`.

Al arrancar se crea la carpeta `data/` con `users.json`, `vehicles.json` y `audit.json`. Está en `.gitignore` y no se sube al repositorio.

La interfaz web (demo visual) está en **http://localhost:8080**. Es una página HTML que solo consume los endpoints de la API listados abajo; no agrega ninguno.

## 5. Endpoints

| Endpoint | HU / RF | Quién | Respuestas |
|---|---|---|---|
| `GET /api/vehicles` | HU-VIS-01 | Cualquiera (sin token) | 200 |
| `POST /api/auth/register` | HU-VEN-01 | Cualquiera | 201, 400, 409 |
| `POST /api/auth/login` | RF-01 Authentication | Cualquiera | 200 (devuelve `token`), 401 |
| `POST /api/vehicles` | HU-VEN-01 | Vendedor | 201 (queda `pendiente`), 400, 401, 403 |
| `PUT /api/vehicles/:id/sold` | HU-VEN-01 | Vendedor dueño | 200, 403, 404, 409 |
| `PUT /api/admin/vehicles/:id/approve` | HU-ADM-01 | Administrador | 200 (pasa a `publicada`), 404, 409 |
| `DELETE /api/admin/vehicles/:id` | HU-ADM-01 | Administrador | 200, 404 |
| `DELETE /api/admin/users/:id` | HU-ADM-01 | Administrador | 200 (borra también sus publicaciones), 404 |

Las rutas protegidas exigen la cabecera `Authorization: Bearer <token>`, con el token que devuelve el login.

## 6. Cómo comprobar el código

Con el servidor corriendo en una ventana, abre una **segunda ventana de PowerShell** y pega primero este ayudante (en PowerShell `curl` es un alias de `Invoke-WebRequest`, por eso se usa `Invoke-RestMethod`):

```powershell
$BASE = "http://localhost:8080"
function Call($Method, $Path, $Token, $Body) {
  $p = @{ Method = $Method; Uri = "$BASE$Path"; ContentType = "application/json; charset=utf-8" }
  if ($Token) { $p.Headers = @{ Authorization = "Bearer $Token" } }
  if ($Body)  { $p.Body = ($Body | ConvertTo-Json) }
  try {
    $global:Last = Invoke-RestMethod @p
    Write-Host "OK" -ForegroundColor Green
    ConvertTo-Json -InputObject $Last -Depth 6
  } catch {
    Write-Host ("HTTP " + [int]$_.Exception.Response.StatusCode) -ForegroundColor Yellow
    $_.ErrorDetails.Message
  }
}
```

### HU-VIS-01 · Catálogo público (sin credenciales)

```powershell
Call GET /api/vehicles
```

Al principio devuelve `[]`.

### HU-VEN-01 · Registro y publicación (estado pendiente)

```powershell
Call POST /api/auth/register $null @{ name = "Ana Gomez"; email = "ana@demo.test"; password = "clave-ana-123" }
Call POST /api/auth/login $null @{ email = "ana@demo.test"; password = "clave-ana-123" }
$tokenAna = $Last.token
Call POST /api/vehicles $tokenAna @{ title = "Renault Koleos 2018"; brand = "Renault"; model = "Koleos"; price = 78000000; description = "Unico dueno" }
$idVeh = $Last.vehicle.id
Call GET /api/vehicles
```

La publicación sale con `"status": "pendiente"` y el catálogo sigue vacío.

### HU-ADM-01 · Aprobación por el administrador

```powershell
Call POST /api/auth/login $null @{ email = "admin@automarket.test"; password = "Admin-demo-2026" }
$tokenAdmin = $Last.token
Call PUT "/api/admin/vehicles/$idVeh/approve" $tokenAdmin
Call GET /api/vehicles
```

Pasa a `"status": "publicada"` y aparece en el catálogo con el nombre y el correo del vendedor. Usa en el login los mismos valores que definiste en las variables de entorno.

### HU-VEN-01 · Reportar la venta

```powershell
Call PUT "/api/vehicles/$idVeh/sold" $tokenAna
Call GET /api/vehicles
```

El vehículo pasa a `vendido` y sale del catálogo.

### HU-ADM-01 · Eliminar publicación o usuario

```powershell
Call DELETE "/api/admin/vehicles/$idVeh" $tokenAdmin
Call DELETE "/api/admin/users/<id-del-usuario>" $tokenAdmin
```

El id del usuario sale de la respuesta del registro (`user.id`). Eliminar un usuario elimina también sus publicaciones.

### AAA · Authentication (debe dar 401)

```powershell
Call POST /api/vehicles $null @{ title = "x"; brand = "x"; model = "x"; price = 1 }
Call PUT "/api/admin/vehicles/$idVeh/approve" "token-inventado"
Call POST /api/auth/login $null @{ email = "ana@demo.test"; password = "incorrecta" }
```

### AAA · Authorization (debe dar 403)

```powershell
Call PUT "/api/admin/vehicles/$idVeh/approve" $tokenAna
Call POST /api/vehicles $tokenAdmin @{ title = "x"; brand = "x"; model = "x"; price = 1 }
```

El vendedor no puede usar rutas de administrador y el administrador no puede publicar. Un segundo vendedor tampoco puede reportar la venta de una publicación ajena (403).

### AAA · Accounting

Cada petición, incluidas las rechazadas con 401, 403 y 404, queda registrada en `data/audit.json` con actor, rol, método, ruta, código de respuesta y un hash encadenado (cada entrada incluye el hash de la anterior).

```powershell
Get-Content data\audit.json
```

- Las peticiones sin token aparecen como `anonimo` / `visitante`.
- Una petición con token válido pero rol incorrecto aparece con el actor real y el código `403`.
- **Integridad:** al arrancar, el servidor recalcula la cadena. Si alguien edita o borra una entrada intermedia, el servidor **se niega a arrancar** con el mensaje `cadena de auditoría alterada: entrada N`. Para probarlo: detén el servidor, cambia un valor en `audit.json` (por ejemplo un `status_response`) y vuelve a arrancar.

Para reiniciar la demo desde cero: detén el servidor y borra la carpeta `data\` (se pierden usuarios, publicaciones y auditoría).

```powershell
Remove-Item -Recurse -Force data
```

## 7. Límites conocidos

- Los hashes de la auditoría **no llevan clave**: detectan ediciones simples, pero no a quien reescriba y recalcule toda la cadena.
- **Truncar las últimas entradas** de `audit.json` tampoco se detecta.
- Si el registro de una operación no se puede escribir, el servidor lo anota en su log y la operación igual se completa.
- Los tokens viven en memoria: no expiran y se pierden al reiniciar el servidor.
- Las contraseñas se guardan con sal y SHA-256 iterado (100.000 rondas), porque bcrypt y librerías similares están prohibidas en el taller.

## 8. Problemas frecuentes

| Síntoma | Causa probable | Solución |
|---|---|---|
| `go : El término 'go' no se reconoce` | Go no está instalado o la terminal se abrió antes de instalarlo | Instala Go y abre una terminal nueva |
| `go mod tidy` falla con `no such host` | Sin internet, proxy o firewall | Revisa la conexión o configura `GOPROXY` |
| `listen tcp :8080: bind` | El puerto 8080 está ocupado | `netstat -ano \| findstr :8080`, luego `taskkill /PID <pid> /F` |
| Error sobre `web/templates/index.html` | Ejecutaste `go run` desde otra carpeta | Entra a la raíz del repo y ejecuta `go run ./cmd/server` |
| `ADMIN_EMAIL/ADMIN_PASSWORD no definidos` | Variables no definidas en esa ventana | Defínelas (sección 3) y reinicia el servidor |
| Login del admin da `401` | Contraseña distinta a la usada al crearlo | Usa la original, o borra `data\` y arranca de nuevo |
| `401 token inválido` de repente | Reiniciaste el servidor: los tokens viven en memoria | Haz login de nuevo |
| El servidor no arranca: `cadena de auditoría alterada` | `audit.json` fue editado o quedó corrupto | Restaura el original o borra `data\audit.json` |
| `Access is denied` al escribir en `data\` | Repo en una carpeta sincronizada (OneDrive) | Pausa la sincronización o mueve el repo |

## 9. Estructura del código

```
cmd/server/          arranque, rutas y semilla del administrador
internal/handlers/   capa HTTP (leer la petición, responder)
internal/service/    reglas de negocio
internal/storage/    persistencia en JSON (con mutex) y cadena de auditoría
internal/middleware/ Authentication, Authorization y Accounting
internal/security/   hash de contraseñas y tokens
internal/httpx/      formato único de error e identidad en el contexto
internal/models/     entidades, constantes y errores de dominio
web/templates/       interfaz HTML (demo visual)
docs/                diagramas de estructura y de comportamiento (Mermaid)
```

## 10. Declaración de uso responsable de IA

El equipo utilizó asistentes de IA (Claude) como apoyo en la revisión del código frente a los requisitos del taller, la limpieza de código, la redacción de pruebas y de la guía de ejecución, y la preparación de la sustentación. Todo el código fue revisado, ejecutado y comprendido por los integrantes, quienes pueden explicarlo y defenderlo. Las decisiones de diseño y los límites de la solución son responsabilidad del equipo. No se incluyeron dependencias distintas de Gin.

Además, la IA se utilizó, en última instancia, para crear el diseño de la interfaz web (`web/templates/index.html`).
