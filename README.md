# GO Pentview Control de Horas

## Requerimientos
- [x] Acceso a todas las rutas con Bearer token (excepto Login)
- [x] Módulos para el rol de Administrador: PERFIL, HORARIO, ROLES y USUARIOS
- [x] Módulos para otros roles: PERFIL y HORARIO
- [x] El Administrador en el módulo de USUARIOS puede: CREAR, LISTAR y ELIMINAR
- [x] Todos los usuarios pueden actualizar sus datos en el módulo PERFIL
- [x] Todos los usuarios pueden registrar ENTRADA/SALIDA en el módulo HORARIO

## Tecnología
- [x] Go

```bash
go run main.go
```

## Dependencias
- github.com/golang-jwt/jwt/v5@v5.2.0
- github.com/gorilla/handlers@v1.5.2
- github.com/gorilla/mux@v1.8.1
- github.com/joho/godotenv@v1.5.1
- github.com/mattn/go-sqlite3@v1.14.19
- golang.org/x/crypto@v0.18.0
- go@1.21.5

## Descripción
Pentview requiere un sistema de Control de Horas para la gestión de su personal permitiendo el registro de horas de entrada/salida y la gestión de sus usuarios con el rol respectivo de cada uno. La plataforma debe contar con controles de la autenticación.

## Solución

Aplicación Backend provisional con las funcionalidades para [Angular Pentview Control de Horas](https://github.com/da8ah/angular-pentview). Cuenta con autenticación y enrutamiento para llevar a cabo las operaciones del Fronted. Las implementaciones son mínimas por lo que pueden faltar varios controles, sin embargo, permite realizar todas las funcionalidades requeridas.

## Versionamiento

(Tiber) **Diciembre 2024 v1.0**

* Actualización de README
* Autenticación
* Funcionalidades Mínimas
* Base de datos e Imágenes
