package service

import "errors"

var (
	ErrForbidden          = errors.New("sin permiso sobre este recurso")
	ErrInvalidCredentials = errors.New("credenciales inválidas")
	ErrCatalogUnavailable = errors.New("no se pudo obtener el contacto del vendedor")
)
