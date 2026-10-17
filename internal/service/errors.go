package service

import "errors"

var (
	ErrForbidden          = errors.New("sin permiso sobre este recurso")
	ErrInvalidCredentials = errors.New("credenciales inválidas")
)
