package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"automarket/internal/storage"

	"github.com/gin-gonic/gin"
)

func TestRegisterHandlerAlwaysCreatesSeller(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store, err := storage.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	body := `{"name":"Ana","email":"ana@example.com","password":"clave-segura","role":"administrador"}`
	request := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router := gin.New()
	router.POST("/api/auth/register", RegisterHandler(store))
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("código HTTP = %d; se esperaba %d", recorder.Code, http.StatusCreated)
	}
	created, err := store.GetUserByEmail("ana@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if created.Role != "vendedor" {
		t.Fatalf("rol asignado = %q; se esperaba vendedor", created.Role)
	}
}
