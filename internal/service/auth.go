package service

import (
	"errors"

	"automarket/internal/models"
	"automarket/internal/security"
)

type authStore interface {
	CreateUser(*models.User) error
	GetUserByEmail(string) (models.User, error)
}

type tokenIssuer interface {
	Create(userID string) (string, error)
}

// AuthService contiene el registro y la autenticación de credenciales (RF-01).
type AuthService struct {
	store  authStore
	tokens tokenIssuer
}

func NewAuthService(store authStore, tokens tokenIssuer) *AuthService {
	return &AuthService{store: store, tokens: tokens}
}

// Register crea siempre una cuenta de vendedor (HU-VEN-01).
func (s *AuthService) Register(name, email, password string) (models.User, error) {
	passwordHash, err := security.HashPassword(password)
	if err != nil {
		return models.User{}, err
	}
	user := models.User{
		Name:         name,
		Email:        email,
		PasswordHash: passwordHash,
		Role:         models.RoleSeller,
	}
	if err := s.store.CreateUser(&user); err != nil {
		return models.User{}, err
	}
	return user, nil
}

// Login verifica credenciales y emite un token Bearer.
func (s *AuthService) Login(email, password string) (models.User, string, error) {
	user, err := s.store.GetUserByEmail(email)
	if errors.Is(err, models.ErrNotFound) {
		security.BurnVerify(password)
		return models.User{}, "", ErrInvalidCredentials
	}
	if err != nil {
		return models.User{}, "", err
	}

	valid, err := security.VerifyPassword(password, user.PasswordHash)
	if err != nil {
		security.BurnVerify(password)
		return models.User{}, "", ErrInvalidCredentials
	}
	if !valid {
		return models.User{}, "", ErrInvalidCredentials
	}
	token, err := s.tokens.Create(user.ID)
	if err != nil {
		return models.User{}, "", err
	}
	return user, token, nil
}
