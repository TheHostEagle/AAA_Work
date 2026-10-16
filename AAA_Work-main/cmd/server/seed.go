package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"automarket/internal/models"
	"automarket/internal/security"
	"automarket/internal/storage"
)

func seedAdmin(store *storage.Store) error {
	email := strings.TrimSpace(os.Getenv("ADMIN_EMAIL"))
	password := os.Getenv("ADMIN_PASSWORD")
	if email == "" || password == "" {
		log.Println("ADMIN_EMAIL/ADMIN_PASSWORD no definidos: no se crea administrador")
		return nil
	}

	user, err := store.GetUserByEmail(email)
	if err == nil {
		if user.Role != models.RoleAdmin {
			return fmt.Errorf("ADMIN_EMAIL ya pertenece a una cuenta que no es administradora")
		}
		return nil
	}
	if !errors.Is(err, storage.ErrNotFound) {
		return err
	}

	hash, err := security.HashPassword(password)
	if err != nil {
		return fmt.Errorf("generando hash del administrador: %w", err)
	}
	admin := models.User{
		Name:         "Administrador",
		Email:        email,
		PasswordHash: hash,
		Role:         models.RoleAdmin,
	}
	if err := store.CreateUser(&admin); err != nil {
		return fmt.Errorf("creando administrador: %w", err)
	}
	return nil
}
