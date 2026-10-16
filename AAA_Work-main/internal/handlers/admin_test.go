package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"automarket/internal/middleware"
	"automarket/internal/models"
	"automarket/internal/security"
	"automarket/internal/storage"

	"github.com/gin-gonic/gin"
)

func newAdminTestStore(t *testing.T) *storage.Store {
	t.Helper()
	store, err := storage.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return store
}

func TestApproveVehicleHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := newAdminTestStore(t)
	vehicle := models.Vehicle{SellerID: "seller-1", Title: "Vehículo pendiente"}
	if err := store.CreateVehicle(&vehicle); err != nil {
		t.Fatal(err)
	}

	router := gin.New()
	router.PUT("/api/admin/vehicles/:id/approve", ApproveVehicleHandler(store))
	request := httptest.NewRequest(http.MethodPut, "/api/admin/vehicles/"+vehicle.ID+"/approve", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("código HTTP = %d; se esperaba %d", recorder.Code, http.StatusOK)
	}
	updated, err := store.GetVehicle(vehicle.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Status != "publicada" {
		t.Fatalf("estado = %q; se esperaba publicada", updated.Status)
	}

	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodPut, "/api/admin/vehicles/"+vehicle.ID+"/approve", nil))
	if recorder.Code != http.StatusConflict {
		t.Fatalf("aprobar vehículo ya publicado: código = %d; se esperaba %d", recorder.Code, http.StatusConflict)
	}

	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodPut, "/api/admin/vehicles/no-existe/approve", nil))
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("vehículo inexistente: código = %d; se esperaba %d", recorder.Code, http.StatusNotFound)
	}
}

func TestDeleteVehicleHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := newAdminTestStore(t)
	vehicle := models.Vehicle{SellerID: "seller-1", Title: "Vehículo para eliminar"}
	if err := store.CreateVehicle(&vehicle); err != nil {
		t.Fatal(err)
	}

	router := gin.New()
	router.DELETE("/api/admin/vehicles/:id", DeleteVehicleHandler(store))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodDelete, "/api/admin/vehicles/"+vehicle.ID, nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("código HTTP = %d; se esperaba %d", recorder.Code, http.StatusOK)
	}
	if _, err := store.GetVehicle(vehicle.ID); err != storage.ErrNotFound {
		t.Fatalf("vehículo sigue disponible o error inesperado: %v", err)
	}

	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodDelete, "/api/admin/vehicles/no-existe", nil))
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("vehículo inexistente: código = %d; se esperaba %d", recorder.Code, http.StatusNotFound)
	}
}

func TestDeleteUserHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := newAdminTestStore(t)
	user := models.User{Name: "Usuario", Email: "usuario@example.com", Role: "vendedor"}
	if err := store.CreateUser(&user); err != nil {
		t.Fatal(err)
	}

	router := gin.New()
	router.DELETE("/api/admin/users/:id", DeleteUserHandler(store))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodDelete, "/api/admin/users/"+user.ID, nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("código HTTP = %d; se esperaba %d", recorder.Code, http.StatusOK)
	}
	if _, err := store.GetUserByID(user.ID); err != storage.ErrNotFound {
		t.Fatalf("usuario sigue disponible o error inesperado: %v", err)
	}

	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodDelete, "/api/admin/users/no-existe", nil))
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("usuario inexistente: código = %d; se esperaba %d", recorder.Code, http.StatusNotFound)
	}
}
func TestAdminEndpointsRequireAuthenticationAndRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := newAdminTestStore(t)
	seller := models.User{Name: "Vendedor", Email: "vendedor@example.com", Role: "vendedor"}
	if err := store.CreateUser(&seller); err != nil {
		t.Fatal(err)
	}
	tokenStore := security.NewTokenStore()
	token, err := tokenStore.Create(seller.ID)
	if err != nil {
		t.Fatal(err)
	}

	router := gin.New()
	adminGroup := router.Group("/api/admin")
	adminGroup.Use(middleware.AuthenticationMiddleware(tokenStore))
	adminGroup.Use(middleware.AuthorizationMiddleware(store, "administrador"))
	adminGroup.DELETE("/vehicles/:id", DeleteVehicleHandler(store))

	unauthenticated := httptest.NewRecorder()
	router.ServeHTTP(unauthenticated, httptest.NewRequest(http.MethodDelete, "/api/admin/vehicles/id-1", nil))
	if unauthenticated.Code != http.StatusUnauthorized {
		t.Fatalf("sin token: código = %d; se esperaba %d", unauthenticated.Code, http.StatusUnauthorized)
	}

	request := httptest.NewRequest(http.MethodDelete, "/api/admin/vehicles/id-1", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	forbidden := httptest.NewRecorder()
	router.ServeHTTP(forbidden, request)
	if forbidden.Code != http.StatusForbidden {
		t.Fatalf("rol vendedor: código = %d; se esperaba %d", forbidden.Code, http.StatusForbidden)
	}
}
