package service

import (
	"errors"
	"testing"

	"automarket/internal/models"
	"automarket/internal/storage"
)

func TestDeleteUserAlsoRemovesTheirVehicles(t *testing.T) {
	store, err := storage.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	seller := models.User{Name: "Ana", Email: "ana@example.com", Role: models.RoleSeller}
	if err := store.CreateUser(&seller); err != nil {
		t.Fatal(err)
	}
	vehicle := models.Vehicle{SellerID: seller.ID, Title: "Koleos", Status: models.StatusPending}
	if err := store.CreateVehicle(&vehicle); err != nil {
		t.Fatal(err)
	}
	if err := NewAdminService(store).DeleteUser(seller.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := store.GetVehicle(vehicle.ID); !errors.Is(err, storage.ErrNotFound) {
		t.Fatalf("el vehículo del usuario eliminado sigue disponible: %v", err)
	}
	if _, err := store.GetUserByID(seller.ID); !errors.Is(err, storage.ErrNotFound) {
		t.Fatalf("el usuario sigue disponible: %v", err)
	}
}
