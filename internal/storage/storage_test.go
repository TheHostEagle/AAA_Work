package storage

import (
	"testing"

	"automarket/internal/models"
)

func TestFlujo(t *testing.T) {
	dir := t.TempDir()
	st, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}
	u := models.User{Name: "Ana", Email: "a@a.com", PasswordHash: "x", Role: "vendedor"}
	if err := st.CreateUser(&u); err != nil || u.ID == "" {
		t.Fatal("crear usuario", err)
	}
	if err := st.CreateUser(&models.User{Email: "a@a.com"}); err != ErrEmailTaken {
		t.Fatal("correo duplicado debía fallar")
	}
	if _, err := st.GetUserByEmail("a@a.com"); err != nil {
		t.Fatal(err)
	}
	v := models.Vehicle{SellerID: u.ID, Title: "Koleos", Price: 1}
	st.CreateVehicle(&v)
	if v.Status != "pendiente" || len(st.ListVehiclesByStatus("publicada")) != 0 {
		t.Fatal("debe nacer pendiente y no verse en catálogo")
	}
	v.Status = "publicada"
	st.UpdateVehicle(v)
	if len(st.ListVehiclesByStatus("publicada")) != 1 {
		t.Fatal("debe verse tras aprobar")
	}
	l1 := models.AuditLog{ActorID: "anonimo", Method: "GET", Endpoint: "/api/vehicles", StatusResponse: 200}
	l2 := models.AuditLog{ActorID: u.ID, Method: "POST", Endpoint: "/api/vehicles", StatusResponse: 201}
	st.SaveAuditLog(&l1)
	st.SaveAuditLog(&l2)
	if l1.PrevHash != "genesis" || l2.PrevHash != l1.Hash || l2.Hash == l1.Hash {
		t.Fatal("cadena de hashes rota")
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
