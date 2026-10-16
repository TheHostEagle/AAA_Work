# Preparar una cuenta administradora para la demostración

El API asigna el rol `vendedor` a todo registro público y no ofrece una ruta para elevar privilegios. Para probar las rutas de administrador en el entorno local del taller, se puede provisionar la cuenta fuera de la API, sin agregar endpoints al alcance.

1. Desde la raíz del proyecto, inicia el servidor con `go run ./cmd/server`.
2. En otra ventana de PowerShell, crea una cuenta temporal con el registro existente. Usa un correo y contraseña de demostración, no credenciales personales:

   ```powershell
   $body = @{ name = 'Admin Demo'; email = 'admin-demo@example.test'; password = 'CAMBIAR-POR-CLAVE-DEMO' } | ConvertTo-Json
   Invoke-RestMethod -Method Post -Uri 'http://localhost:8080/api/auth/register' -ContentType 'application/json' -Body $body
   ```

3. Detén el servidor con `Ctrl+C`. Esto es necesario porque el proceso mantiene los usuarios cargados en memoria.
4. En `data/users.json`, localiza solo la cuenta `admin-demo@example.test` y cambia su campo `role` de `vendedor` a `administrador`. Conserva intactos el ID y `password_hash` generados por la aplicación.
5. Reinicia el servidor. Inicia sesión con esa cuenta y usa el token devuelto en `Authorization: Bearer <token>` para las rutas `/api/admin/...`.
6. Al terminar la demostración, elimina los datos locales de prueba. No incluyas `data/users.json`, contraseñas de demostración ni tokens en el repositorio o en el PDF.

Esta preparación es manual y exclusiva del entorno de demostración. No añade una API para cambiar roles ni altera el flujo público de registro.
