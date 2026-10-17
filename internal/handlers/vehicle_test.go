package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"automarket/internal/httpx"
	"automarket/internal/middleware"
	"automarket/internal/models"
	"automarket/internal/security"
	"automarket/internal/service"
	"automarket/internal/storage"

	"github.com/gin-gonic/gin"
)

func newVehicleTestStore(t *testing.T) (*storage.Store, models.User) {
	t.Helper()
	store, err := storage.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	seller := models.User{Name: "Ana Vendedora", Email: "ana@example.com", PasswordHash: "secreto-hasheado", Role: models.RoleSeller}
	if err := store.CreateUser(&seller); err != nil {
		t.Fatal(err)
	}
	return store, seller
}

func TestGetPublicVehiclesHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store, seller := newVehicleTestStore(t)

	published := models.Vehicle{SellerID: seller.ID, Title: "Sedán publicado", Price: 12000, Status: models.StatusPending}
	if err := store.CreateVehicle(&published); err != nil {
		t.Fatal(err)
	}
	if _, err := store.ChangeVehicleStatus(published.ID, models.StatusPending, models.StatusPublished); err != nil {
		t.Fatal(err)
	}

	pending := models.Vehicle{SellerID: seller.ID, Title: "Vehículo pendiente", Price: 9000, Status: models.StatusPending}
	if err := store.CreateVehicle(&pending); err != nil {
		t.Fatal(err)
	}

	router := gin.New()
	router.GET("/api/vehicles", GetPublicVehiclesHandler(service.NewVehicleService(store)))
	request := httptest.NewRequest(http.MethodGet, "/api/vehicles", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("código HTTP = %d; se esperaba %d", recorder.Code, http.StatusOK)
	}

	var got []models.PublicVehicle
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
	router.GET("/api/vehicles", GetPublicVehiclesHandler(service.NewVehicleService(store)))

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/vehicles", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("código HTTP = %d; se esperaba %d", recorder.Code, http.StatusOK)
	}
	if strings.TrimSpace(recorder.Body.String()) != "[]" {
		t.Fatalf("respuesta = %q; se esperaba []", recorder.Body.String())
	}

}

func asUser(userID string) gin.HandlerFunc {
	return func(c *gin.Context) {
		httpx.SetUserID(c, userID)
		c.Next()
	}
}

func sendJSON(router *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

func createVehicle(t *testing.T, store *storage.Store, sellerID, status string) models.Vehicle {
	t.Helper()
	vehicle := models.Vehicle{SellerID: sellerID, Title: "Vehículo " + status, Status: models.StatusPending}
	if err := store.CreateVehicle(&vehicle); err != nil {
		t.Fatal(err)
	}
	if status != models.StatusPending {
		if _, err := store.ChangeVehicleStatus(vehicle.ID, models.StatusPending, status); err != nil {
			t.Fatal(err)
		}
	}
	return vehicle
}

//publicar vehiculo

func TestCreateVehicleHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store, seller := newVehicleTestStore(t)
	router := gin.New()
	vehicles := service.NewVehicleService(store)
	router.POST("/api/vehicles", asUser(seller.ID), CreateVehicleHandler(vehicles))
	body := `{"title":"Koleos 2018","brand":"Renault","model":"Koleos","price":61900000,` +
		`"status":"publicada","seller_id":"otro-usuario"}`
	recorder := sendJSON(router, http.MethodPost, "/api/vehicles", body)
	if recorder.Code != http.StatusCreated {
		t.Fatalf("código HTTP = %d; se esperaba %d", recorder.Code, http.StatusCreated)
	}
	pending := store.ListVehiclesByStatus(models.StatusPending)
	if len(pending) != 1 {
		t.Fatalf("publicaciones pendientes = %d; se esperaba 1", len(pending))
	}
	if pending[0].SellerID != seller.ID {
		t.Fatalf("seller_id = %q; se esperaba %q", pending[0].SellerID, seller.ID)
	}
	if len(store.ListVehiclesByStatus(models.StatusPublished)) != 0 {
		t.Fatal("una publicación nueva no debe verse en el catálogo")
	}
}

func TestCreateVehicleHandlerRejectsInvalidData(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store, seller := newVehicleTestStore(t)
	router := gin.New()
	router.POST("/api/vehicles", asUser(seller.ID), CreateVehicleHandler(service.NewVehicleService(store)))
	invalidBodies := []string{
		`{"title":"","brand":"Renault","model":"Koleos","price":1000}`,
		`{"title":"Koleos","brand":"Renault","model":"Koleos","price":0}`,
		`esto no es json`,
	}
	for _, body := range invalidBodies {
		recorder := sendJSON(router, http.MethodPost, "/api/vehicles", body)
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("cuerpo %q: código = %d; se esperaba %d",
				body, recorder.Code, http.StatusBadRequest)
		}
	}
	if len(store.ListVehiclesByStatus(models.StatusPending)) != 0 {
		t.Fatal("no debe guardarse ninguna publicación inválida")
	}
}

func TestMarkVehicleAsSoldHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store, seller := newVehicleTestStore(t)
	other := models.User{Name: "Otro", Email: "otro@example.com", Role: models.RoleSeller}
	if err := store.CreateUser(&other); err != nil {
		t.Fatal(err)
	}
	published := createVehicle(t, store, seller.ID, models.StatusPublished)
	pending := createVehicle(t, store, seller.ID, models.StatusPending)
	vehicles := service.NewVehicleService(store)
	ownerRouter := gin.New()
	ownerRouter.PUT("/api/vehicles/:id/sold", asUser(seller.ID), MarkVehicleAsSoldHandler(vehicles))
	otherRouter := gin.New()
	otherRouter.PUT("/api/vehicles/:id/sold", asUser(other.ID), MarkVehicleAsSoldHandler(vehicles))
	recorder := sendJSON(otherRouter, http.MethodPut, "/api/vehicles/"+published.ID+"/sold", "")
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("publicación ajena: código = %d; se esperaba %d", recorder.Code, http.StatusForbidden)
	}
	recorder = sendJSON(ownerRouter, http.MethodPut, "/api/vehicles/"+pending.ID+"/sold", "")
	if recorder.Code != http.StatusConflict {
		t.Fatalf("publicación pendiente: código = %d; se esperaba %d", recorder.Code, http.StatusConflict)
	}
	recorder = sendJSON(ownerRouter, http.MethodPut, "/api/vehicles/no-existe/sold", "")
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("vehículo inexistente: código = %d; se esperaba %d", recorder.Code, http.StatusNotFound)
	}
	recorder = sendJSON(ownerRouter, http.MethodPut, "/api/vehicles/"+published.ID+"/sold", "")
	if recorder.Code != http.StatusOK {
		t.Fatalf("dueño: código = %d; se esperaba %d", recorder.Code, http.StatusOK)
	}
	updated, err := store.GetVehicle(published.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Status != models.StatusSold {
		t.Fatalf("estado = %q; se esperaba vendido", updated.Status)
	}
	if len(store.ListVehiclesByStatus(models.StatusPublished)) != 0 {
		t.Fatal("un vehículo vendido no debe verse en el catálogo")
	}
	recorder = sendJSON(ownerRouter, http.MethodPut, "/api/vehicles/"+published.ID+"/sold", "")
	if recorder.Code != http.StatusConflict {
		t.Fatalf("venta repetida: código = %d; se esperaba %d", recorder.Code, http.StatusConflict)
	}
}

func TestSellerEndpointsRequireSellerRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store, seller := newVehicleTestStore(t)
	admin := models.User{Name: "Admin", Email: "admin@example.com", Role: models.RoleAdmin}
	if err := store.CreateUser(&admin); err != nil {
		t.Fatal(err)
	}
	tokenStore := security.NewTokenStore()
	sellerToken, err := tokenStore.Create(seller.ID)
	if err != nil {
		t.Fatal(err)
	}
	adminToken, err := tokenStore.Create(admin.ID)
	if err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	sellerGroup := router.Group("/api/vehicles")
	sellerGroup.Use(middleware.AuthenticationMiddleware(tokenStore))
	sellerGroup.Use(middleware.AuthorizationMiddleware(store, models.RoleSeller))
	sellerGroup.POST("", CreateVehicleHandler(service.NewVehicleService(store)))
	body := `{"title":"Koleos","brand":"Renault","model":"Koleos","price":1000}`
	postWithToken := func(token string) int {
		request := httptest.NewRequest(http.MethodPost, "/api/vehicles", strings.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		if token != "" {
			request.Header.Set("Authorization", "Bearer "+token)
		}
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)
		return recorder.Code
	}
	if code := postWithToken(""); code != http.StatusUnauthorized {
		t.Fatalf("sin token: código = %d; se esperaba %d", code, http.StatusUnauthorized)
	}
	if code := postWithToken(adminToken); code != http.StatusForbidden {
		t.Fatalf("rol administrador: código = %d; se esperaba %d", code, http.StatusForbidden)
	}
	if code := postWithToken(sellerToken); code != http.StatusCreated {
		t.Fatalf("rol vendedor: código = %d; se esperaba %d", code, http.StatusCreated)
	}
}
