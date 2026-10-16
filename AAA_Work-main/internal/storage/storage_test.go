package storage

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"automarket/internal/models"
)

func TestFlujo(t *testing.T) {
	dir := t.TempDir()
	st, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}
	u := models.User{Name: "Ana", Email: "a@a.com", PasswordHash: "x", Role: models.RoleSeller}
	if err := st.CreateUser(&u); err != nil || u.ID == "" {
		t.Fatal("crear usuario", err)
	}
	if err := st.CreateUser(&models.User{Email: "a@a.com"}); err != ErrEmailTaken {
		t.Fatal("correo duplicado debía fallar")
	}
	if _, err := st.GetUserByEmail("a@a.com"); err != nil {
		t.Fatal(err)
	}
	v := models.Vehicle{SellerID: u.ID, Title: "Koleos", Price: 1, Status: models.StatusPending}
	if err := st.CreateVehicle(&v); err != nil {
		t.Fatal(err)
	}
	if v.Status != models.StatusPending || len(st.ListVehiclesByStatus(models.StatusPublished)) != 0 {
		t.Fatal("debe nacer pendiente y no verse en catálogo")
	}
	if _, err := st.ChangeVehicleStatus(v.ID, models.StatusPending, models.StatusPublished); err != nil {
		t.Fatal(err)
	}
	if len(st.ListVehiclesByStatus(models.StatusPublished)) != 1 {
		t.Fatal("debe verse tras aprobar")
	}
	l1 := models.AuditLog{ActorID: "anonimo", Method: "GET", Endpoint: "/api/vehicles", StatusResponse: 200}
	l2 := models.AuditLog{ActorID: u.ID, Method: "POST", Endpoint: "/api/vehicles", StatusResponse: 201}
	st.SaveAuditLog(&l1)
	st.SaveAuditLog(&l2)
	if l1.PrevHash != "genesis" || l2.PrevHash != l1.Hash || l2.Hash == l1.Hash {
		t.Fatal("cadena de hashes rota")
	}
	if err := st.VerifyAuditChain(); err != nil {
		t.Fatalf("cadena de auditoría válida rechazada: %v", err)
	}
	// persistencia: reabrir y comprobar que sigue todo
	st2, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := st2.GetVehicle(v.ID); err != nil {
		t.Fatal("no persistió el vehículo")
	}
	if err := st2.DeleteVehicle(v.ID); err != nil {
		t.Fatal(err)
	}
	if err := st2.DeleteUser(u.ID); err != nil {
		t.Fatal(err)
	}
}

func TestVerifyAuditChainDetectsTampering(t *testing.T) {
	st, err := New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	entry := models.AuditLog{Method: "GET", Endpoint: "/api/vehicles", StatusResponse: 200}
	if err := st.SaveAuditLog(&entry); err != nil {
		t.Fatal(err)
	}
	st.logs[0].Endpoint = "/api/admin/users"
	if err := st.VerifyAuditChain(); !errors.Is(err, ErrAuditTampered) {
		t.Fatalf("error = %v; se esperaba ErrAuditTampered", err)
	}
}

func TestFailedCreateDoesNotChangeMemory(t *testing.T) {
	dir := t.TempDir()
	st, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "users.json"), 0o700); err != nil {
		t.Fatal(err)
	}
	user := models.User{Email: "ana@example.com"}
	if err := st.CreateUser(&user); err == nil {
		t.Fatal("se esperaba un error al guardar sobre un directorio")
	}
	if _, err := st.GetUserByEmail(user.Email); !errors.Is(err, ErrNotFound) {
		t.Fatalf("el usuario fallido quedó en memoria: %v", err)
	}
}
