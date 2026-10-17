package models

import "errors"

// Errores de dominio compartidos por storage, service y handlers.
var (
	ErrNotFound       = errors.New("registro no encontrado")
	ErrEmailTaken     = errors.New("el correo ya está registrado")
	ErrStatusMismatch = errors.New("estado actual incompatible con la operación")
)
