package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"automarket/internal/models"
	"automarket/internal/security"
	"automarket/internal/storage"

	"github.com/gin-gonic/gin"
)

func TestAccountingMiddlewareRecordsEveryRequestInHashChain(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dir := t.TempDir()
	store, err := storage.New(dir)
	if err != nil {
		t.Fatal(err)
	}

	user := models.User{Name: "Vendedor", Email: "seller@example.com", Role: "vendedor"}
	if err := store.CreateUser(&user); err != nil {
		t.Fatal(err)
	}
	tokens := security.NewTokenStore()
	token, err := tokens.Create(user.ID)
	if err != nil {
		t.Fatal(err)
	}

	router := gin.New()
	router.Use(AccountingMiddleware(store))
	router.GET("/public/vehicles/:id", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
	router.GET("/seller", AuthenticationMiddleware(tokens), func(c *gin.Context) {
		c.Status(http.StatusAccepted)
	})
	router.GET("/private", AuthenticationMiddleware(tokens), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	publicResponse := httptest.NewRecorder()
	router.ServeHTTP(publicResponse, httptest.NewRequest(http.MethodGet, "/public/vehicles/vehicle-1", nil))
	if publicResponse.Code != http.StatusNoContent {
		t.Fatalf("catálogo público: código = %d", publicResponse.Code)
	}

	sellerRequest := httptest.NewRequest(http.MethodGet, "/seller", nil)
	sellerRequest.Header.Set("Authorization", "Bearer "+token)
	sellerResponse := httptest.NewRecorder()
	router.ServeHTTP(sellerResponse, sellerRequest)
	if sellerResponse.Code != http.StatusAccepted {
		t.Fatalf("petición autenticada: código = %d", sellerResponse.Code)
	}

	unauthenticatedResponse := httptest.NewRecorder()
	router.ServeHTTP(unauthenticatedResponse, httptest.NewRequest(http.MethodGet, "/private", nil))
	if unauthenticatedResponse.Code != http.StatusUnauthorized {
		t.Fatalf("petición sin token: código = %d", unauthenticatedResponse.Code)
	}

	data, err := os.ReadFile(filepath.Join(dir, "audit.json"))
	if err != nil {
		t.Fatal(err)
	}
	var entries []models.AuditLog
	if err := json.Unmarshal(data, &entries); err != nil {
		t.Fatal(err)
	}
	if len(entries) != 3 {
		t.Fatalf("registros de auditoría = %d; se esperaban 3", len(entries))
	}
	if entries[0].ActorID != "anonimo" || entries[0].ActorRole != "visitante" ||
		entries[0].Endpoint != "/public/vehicles/vehicle-1" || entries[0].StatusResponse != http.StatusNoContent {
		t.Fatalf("registro público inesperado: %+v", entries[0])
	}
	if entries[1].ActorID != user.ID || entries[1].ActorRole != "vendedor" ||
		entries[1].StatusResponse != http.StatusAccepted {
		t.Fatalf("registro autenticado inesperado: %+v", entries[1])
	}
	if entries[2].StatusResponse != http.StatusUnauthorized {
		t.Fatalf("registro de rechazo inesperado: %+v", entries[2])
	}
	if entries[0].PrevHash != "genesis" || entries[1].PrevHash != entries[0].Hash ||
		entries[2].PrevHash != entries[1].Hash || entries[0].Hash == "" || entries[2].Hash == "" {
		t.Fatal("la cadena de hashes de auditoría está rota")
	}
}
