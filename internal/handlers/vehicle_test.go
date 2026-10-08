package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"automarket/internal/models"
	"automarket/internal/storage"

	"github.com/gin-gonic/gin"
)

func newVehicleTestStore(t *testing.T) (*storage.Store, models.User) {
	t.Helper()
	store, err := storage.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	seller := models.User{Name: "Ana Vendedora", Email: "ana@example.com", PasswordHash: "secreto-hasheado", Role: "vendedor"}
	if err := store.CreateUser(&seller); err != nil {
		t.Fatal(err)
	}
	return store, seller
}

func TestGetPublicVehiclesHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store, seller := newVehicleTestStore(t)

	published := models.Vehicle{SellerID: seller.ID, Title: "Sedán publicado", Price: 12000}
	if err := store.CreateVehicle(&published); err != nil {
		t.Fatal(err)
	}
	published.Status = "publicada"
	if err := store.UpdateVehicle(published); err != nil {
		t.Fatal(err)
	}

	pending := models.Vehicle{SellerID: seller.ID, Title: "Vehículo pendiente", Price: 9000}
	if err := store.CreateVehicle(&pending); err != nil {
		t.Fatal(err)
	}

	router := gin.New()
	router.GET("/api/vehicles", GetPublicVehiclesHandler(store))
	request := httptest.NewRequest(http.MethodGet, "/api/vehicles", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("código HTTP = %d; se esperaba %d", recorder.Code, http.StatusOK)
	}

	var got []publicVehicle
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatalf("respuesta JSON inválida: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("cantidad de vehículos = %d; se esperaba 1", len(got))
	}
	if got[0].ID != published.ID || got[0].Seller.Name != seller.Name || got[0].Seller.Email != seller.Email {
		t.Fatalf("respuesta inesperada: %+v", got[0])
	}
	if strings.Contains(recorder.Body.String(), "password_hash") || strings.Contains(recorder.Body.String(), "secreto-hasheado") {
		t.Fatal("la respuesta pública expuso datos de contraseña")
	}
}

func TestGetPublicVehiclesHandlerReturnsEmptyArray(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store, _ := newVehicleTestStore(t)
	router := gin.New()
	router.GET("/api/vehicles", GetPublicVehiclesHandler(store))

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/vehicles", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("código HTTP = %d; se esperaba %d", recorder.Code, http.StatusOK)
	}
	if strings.TrimSpace(recorder.Body.String()) != "[]" {
		t.Fatalf("respuesta = %q; se esperaba []", recorder.Body.String())
	}
}
