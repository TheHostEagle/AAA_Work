# Revisión del taller AAA

## Qué implementa el repositorio

| Área | Implementación localizada | Rutas |
|---|---|---|
| Registro e inicio de sesión | `internal/handlers/auth.go`; contraseña con salt aleatorio y SHA-256 iterado; tokens aleatorios en memoria | `POST /api/auth/register`, `POST /api/auth/login` |
| Catálogo público | Solo vehículos en estado `publicada`; respuesta incluye nombre y email del vendedor | `GET /api/vehicles` |
| Vendedor | Requiere token Bearer y rol `vendedor`; crea publicación `pendiente` asociada al usuario; el propietario puede marcarla `vendido` | `POST /api/vehicles`, `PUT /api/vehicles/:id/sold` |
| Administrador | Requiere token Bearer y rol `administrador`; aprueba pendientes y elimina vehículos o cuentas | `PUT /api/admin/vehicles/:id/approve`, `DELETE /api/admin/vehicles/:id`, `DELETE /api/admin/users/:id` |
| Accounting | Middleware global registra método, ruta, actor, rol y estado HTTP en `audit.json`; los hashes incluyen el hash previo | Se aplica a las rutas API registradas |
| Persistencia | Archivos JSON locales, protegidos por mutex | `users.json`, `vehicles.json`, `audit.json` en `data/` |

## Diagramas

El diagrama previo omitía el flujo de autenticación, las dos rutas del vendedor, el middleware global de accounting y el archivo de auditoría. Se reemplazó por `diagrama-estructura.mmd`, que refleja esos componentes y los tres archivos usados por el Store. `diagrama-comportamiento.mmd` es el flujo final preparado desde el comportamiento visible en handlers y rutas; el equipo puede revisarlo contra el guion de la sustentación.

## Pendientes y puntos a resolver antes de entregar

1. **Cuenta administradora para la demostración:** el registro público asigna siempre el rol `vendedor` y no existe ruta de alta de administradores. Se documentó una preparación local para la demostración editando el rol en `data/users.json`, sin agregar ruta ni capacidad a la API. Consultar `provisionar-admin.md`.
2. **Datos del PDF:** faltan nombres de integrantes, URL del repositorio, hash del commit final, URL del video, confirmación de invitación a `imcabezas` y aprobación del equipo para la declaración de uso responsable de IA.
3. **Declaración de alcance del token:** `TokenStore` es volátil; reiniciar el proceso invalida todos los tokens. Es compatible con la persistencia local mostrada, pero conviene explicarlo en la sustentación.
4. **Garantía de Accounting:** el middleware reporta el fallo de escritura en el log del servidor y permite que la operación HTTP termine. Si el archivo no puede escribirse, no queda garantizado el registro persistente que pide RF-01; el equipo debe decidir si ese fallo debe impedir responder como operación exitosa.
5. **Hash de contraseñas:** la rutina propia usa SHA-256 iterado porque se prohíben dependencias adicionales. No debe describirse como bcrypt ni como PBKDF2/Argon2. La verificación actual no acota los valores de iteraciones recuperados del archivo; si se alterara un hash guardado podría causar consumo excesivo de CPU.
6. **Estado de Git:** la carpeta de trabajo disponible no contiene un commit resoluble ni un remoto configurado; por ello no se puede afirmar un hash de entrega ni URL de repositorio a partir de este checkout.

## Declaración de uso responsable de IA

Texto sugerido para que el equipo lo revise y apruebe antes de incluirlo: “Usamos herramientas de IA como apoyo para revisar el alcance, analizar código y mejorar la documentación y los diagramas. El equipo verificó los resultados contra las historias de usuario, los requisitos y la implementación, y asume la responsabilidad por el código y la entrega final. No se usó IA para atribuir trabajo no realizado ni para exponer datos personales o credenciales.”

