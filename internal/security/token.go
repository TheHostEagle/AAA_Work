package security

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
)

const tokenBytes = 32

var ErrInvalidToken = errors.New("token inválido")

// TokenStore guarda qué usuario pertenece a cada token.
type TokenStore struct {
	mu     sync.RWMutex
	tokens map[string]string
}

// NewTokenStore crea un almacén vacío de tokens.
func NewTokenStore() *TokenStore {
	return &TokenStore{
		tokens: make(map[string]string),
	}
}

// GenerateToken genera un token aleatorio criptográficamente seguro.
func GenerateToken() (string, error) {
	b := make([]byte, tokenBytes)

	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}

// Create genera un token y lo relaciona con un usuario.
func (s *TokenStore) Create(userID string) (string, error) {
	token, err := GenerateToken()
	if err != nil {
		return "", err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.tokens[token] = userID

	return token, nil
}

// GetUserID busca qué usuario pertenece a un token.
func (s *TokenStore) GetUserID(token string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	userID, ok := s.tokens[token]

	return userID, ok
}

// Delete elimina un token del almacén.
func (s *TokenStore) Delete(token string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.tokens, token)
}