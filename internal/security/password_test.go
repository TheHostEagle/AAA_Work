package security

import (
	"errors"
	"testing"
)

func TestVerifyPasswordAcceptsOwnHash(t *testing.T) {
	hash, err := HashPassword("clave-segura")
	if err != nil {
		t.Fatal(err)
	}
	if ok, err := VerifyPassword("clave-segura", hash); err != nil || !ok {
		t.Fatalf("ok=%v err=%v; se esperaba contraseña válida", ok, err)
	}
	if ok, _ := VerifyPassword("otra", hash); ok {
		t.Fatal("una contraseña incorrecta fue aceptada")
	}
}

func TestVerifyPasswordRejectsOutOfRangeRounds(t *testing.T) {
	for _, encoded := range []string{
		"0$00$00",
		"-5$00$00",
		"2000000000$00$00",
	} {
		if _, err := VerifyPassword("x", encoded); !errors.Is(err, ErrInvalidHash) {
			t.Fatalf("%q: error = %v; se esperaba ErrInvalidHash", encoded, err)
		}
	}
}
