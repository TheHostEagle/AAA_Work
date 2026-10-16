package main

import (
	"errors"
	"testing"

	"automarket/internal/models"
	"automarket/internal/storage"
)

func TestSeedAdminCreatesConfiguredAdministrator(t *testing.T) {
	store, err := storage.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("ADMIN_EMAIL", "admin@example.com")
	t.Setenv("ADMIN_PASSWORD", "demo-segura")
	if err := seedAdmin(store); err != nil {
		t.Fatal(err)
	}
	admin, err := store.GetUserByEmail("admin@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if admin.Role != models.RoleAdmin || admin.PasswordHash == "demo-segura" {
		t.Fatalf("administrador mal creado: %+v", admin)
	}
}

func TestSeedAdminDoesNotPromoteExistingSeller(t *testing.T) {
	store, err := storage.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	seller := models.User{Name: "Vendedor", Email: "seller@example.com", Role: models.RoleSeller}
	if err := store.CreateUser(&seller); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ADMIN_EMAIL", seller.Email)
	t.Setenv("ADMIN_PASSWORD", "demo-segura")
	if err := seedAdmin(store); err == nil {
		t.Fatal("se esperaba error cuando ADMIN_EMAIL pertenece a vendedor")
	}
	got, err := store.GetUserByID(seller.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Role != models.RoleSeller {
		t.Fatalf("rol cambiado inesperadamente a %q", got.Role)
	}
}

func TestSeedAdminSkipsWhenEnvironmentIsMissing(t *testing.T) {
	store, err := storage.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("ADMIN_EMAIL", "")
	t.Setenv("ADMIN_PASSWORD", "")
	if err := seedAdmin(store); err != nil {
		t.Fatal(err)
	}
	if _, err := store.GetUserByEmail("admin@example.com"); !errors.Is(err, storage.ErrNotFound) {
		t.Fatalf("administrador inesperado: %v", err)
	}
}
