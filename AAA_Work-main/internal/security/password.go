package security

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Cuántas veces repetimos el hash. Mientras más vueltas, más le toma a
// alguien adivinar la contraseña a la fuerza si se roba la base de datos.
const rounds = 100_000

// Tamaño del valor random que se le suma a cada contraseña. Esto evita
// que dos personas con la misma contraseña terminen con el mismo hash
// guardado.
const saltBytes = 16

var ErrInvalidHash = errors.New("formato de hash inválido")

// HashPassword recibe la contraseña en texto plano y devuelve algo seguro
// para guardar en la base de datos. Nunca guardamos la contraseña real.
func HashPassword(password string) (string, error) {
	salt := make([]byte, saltBytes)

	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generando salt: %w", err)
	}

	hash := derive(password, salt, rounds)

	// Guardamos todo junto en un solo texto: vueltas, salt y hash,
	// separados por "$". Así después podemos volver a verificarlo.
	encoded := fmt.Sprintf(
		"%d$%s$%s",
		rounds,
		hex.EncodeToString(salt),
		hex.EncodeToString(hash),
	)

	return encoded, nil
}

// VerifyPassword revisa si una contraseña escrita por el usuario coincide
// con el hash que ya teníamos guardado (por ejemplo, al hacer login).
func VerifyPassword(password, encoded string) (bool, error) {
	r, salt, wantHash, err := decode(encoded)

	if err != nil {
		return false, err
	}

	gotHash := derive(password, salt, r)

	// Comparamos los hashes usando ConstantTimeCompare.
	return subtle.ConstantTimeCompare(gotHash, wantHash) == 1, nil
}

// derive es el que realmente hace el trabajo pesado: le aplica SHA-256 a la
// contraseña una y otra vez, muchas veces seguidas, para que calcularlo
// sea lento a propósito.
func derive(password string, salt []byte, r int) []byte {
	data := append([]byte(password), salt...)

	sum := sha256.Sum256(data)
	h := sum[:]

	for i := 1; i < r; i++ {
		sum = sha256.Sum256(h)
		h = sum[:]
	}

	return h
}

// decode separa el texto guardado de vuelta en sus 3 partes:
// vueltas, salt y hash.
func decode(encoded string) (r int, salt, hash []byte, err error) {
	parts := strings.Split(encoded, "$")

	if len(parts) != 3 {
		return 0, nil, nil, ErrInvalidHash
	}

	r, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, nil, nil, ErrInvalidHash
	}

	salt, err = hex.DecodeString(parts[1])
	if err != nil {
		return 0, nil, nil, ErrInvalidHash
	}

	hash, err = hex.DecodeString(parts[2])
	if err != nil {
		return 0, nil, nil, ErrInvalidHash
	}

	return r, salt, hash, nil
}