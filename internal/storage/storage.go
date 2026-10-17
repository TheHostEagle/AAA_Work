// Package storage implementa la persistencia del proyecto en archivos JSON.
package storage

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"automarket/internal/models"
)

// ErrAuditTampered indica que la cadena de hashes de auditoría fue alterada.
var ErrAuditTampered = errors.New("cadena de auditoría alterada")

// Store protege con un mutex los registros en memoria y sus archivos JSON.
type Store struct {
	mu       sync.Mutex
	dir      string
	users    []models.User
	vehicles []models.Vehicle
	logs     []models.AuditLog
}

// New crea la carpeta de datos y carga los registros guardados.
func New(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	s := &Store{dir: dir}
	if err := s.load("users.json", &s.users); err != nil {
		return nil, err
	}
	if err := s.load("vehicles.json", &s.vehicles); err != nil {
		return nil, err
	}
	if err := s.load("audit.json", &s.logs); err != nil {
		return nil, err
	}
	return s, nil
}

// CreateUser asigna ID y fecha de creación, y luego persiste al nuevo usuario.
func (s *Store) CreateUser(user *models.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, existing := range s.users {
		if existing.Email == user.Email {
			return models.ErrEmailTaken
		}
	}

	created := *user
	id, err := newID()
	if err != nil {
		return err
	}
	created.ID = id
	created.CreatedAt = now()
	next := append(append([]models.User(nil), s.users...), created)
	if err := s.save("users.json", next); err != nil {
		return err
	}
	s.users = next
	*user = created
	return nil
}

func (s *Store) GetUserByEmail(email string) (models.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, user := range s.users {
		if user.Email == email {
			return user, nil
		}
	}
	return models.User{}, models.ErrNotFound
}

func (s *Store) GetUserByID(id string) (models.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, user := range s.users {
		if user.ID == id {
			return user, nil
		}
	}
	return models.User{}, models.ErrNotFound
}

// DeleteUserWithVehicles elimina al usuario y todas sus publicaciones bajo un
// solo lock. Las publicaciones se persisten primero: si el segundo guardado
// falla, el usuario sigue existiendo y la operación puede repetirse sin dejar
// publicaciones huérfanas.
func (s *Store) DeleteUserWithVehicles(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	userIndex := -1
	for i, user := range s.users {
		if user.ID == id {
			userIndex = i
			break
		}
	}
	if userIndex < 0 {
		return models.ErrNotFound
	}

	remainingVehicles := make([]models.Vehicle, 0, len(s.vehicles))
	for _, vehicle := range s.vehicles {
		if vehicle.SellerID != id {
			remainingVehicles = append(remainingVehicles, vehicle)
		}
	}
	if len(remainingVehicles) != len(s.vehicles) {
		if err := s.save("vehicles.json", remainingVehicles); err != nil {
			return err
		}
		s.vehicles = remainingVehicles
	}

	remainingUsers := append(append([]models.User(nil), s.users[:userIndex]...), s.users[userIndex+1:]...)
	if err := s.save("users.json", remainingUsers); err != nil {
		return err
	}
	s.users = remainingUsers
	return nil
}

// CreateVehicle asigna ID y fecha de creación. El servicio define el estado
// inicial y el dueño antes de llamar a este método de persistencia.
func (s *Store) CreateVehicle(vehicle *models.Vehicle) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	created := *vehicle
	id, err := newID()
	if err != nil {
		return err
	}
	created.ID = id
	created.CreatedAt = now()
	next := append(append([]models.Vehicle(nil), s.vehicles...), created)
	if err := s.save("vehicles.json", next); err != nil {
		return err
	}
	s.vehicles = next
	*vehicle = created
	return nil
}

func (s *Store) GetVehicle(id string) (models.Vehicle, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, vehicle := range s.vehicles {
		if vehicle.ID == id {
			return vehicle, nil
		}
	}
	return models.Vehicle{}, models.ErrNotFound
}

func (s *Store) ListVehiclesByStatus(status string) []models.Vehicle {
	s.mu.Lock()
	defer s.mu.Unlock()
	vehicles := make([]models.Vehicle, 0)
	for _, vehicle := range s.vehicles {
		if vehicle.Status == status {
			vehicles = append(vehicles, vehicle)
		}
	}
	return vehicles
}

// ChangeVehicleStatus verifica el estado actual y persiste la transición bajo
// un solo lock, para que transiciones concurrentes no se pisen entre sí.
func (s *Store) ChangeVehicleStatus(id, from, to string) (models.Vehicle, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, vehicle := range s.vehicles {
		if vehicle.ID != id {
			continue
		}
		if vehicle.Status != from {
			return models.Vehicle{}, models.ErrStatusMismatch
		}
		next := append([]models.Vehicle(nil), s.vehicles...)
		next[i].Status = to
		if err := s.save("vehicles.json", next); err != nil {
			return models.Vehicle{}, err
		}
		s.vehicles = next
		return next[i], nil
	}
	return models.Vehicle{}, models.ErrNotFound
}

func (s *Store) DeleteVehicle(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, vehicle := range s.vehicles {
		if vehicle.ID != id {
			continue
		}
		next := append([]models.Vehicle(nil), s.vehicles[:i]...)
		next = append(next, s.vehicles[i+1:]...)
		if err := s.save("vehicles.json", next); err != nil {
			return err
		}
		s.vehicles = next
		return nil
	}
	return models.ErrNotFound
}

// SaveAuditLog completa los campos de la cadena y agrega la entrada solo después
// de que el archivo completo se haya guardado correctamente.
func (s *Store) SaveAuditLog(entry *models.AuditLog) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	created := *entry
	id, err := newID()
	if err != nil {
		return err
	}
	created.ID = id
	created.Timestamp = now()
	created.PrevHash = "genesis"
	if len(s.logs) > 0 {
		created.PrevHash = s.logs[len(s.logs)-1].Hash
	}
	created.Hash = hashLog(created)
	next := append(append([]models.AuditLog(nil), s.logs...), created)
	if err := s.save("audit.json", next); err != nil {
		return err
	}
	s.logs = next
	*entry = created
	return nil
}

// VerifyAuditChain detecta enlaces rotos o entradas cuyo contenido guardado ya
// no coincide con su hash. No impide que un atacante reconstruya toda la
// cadena, porque los hashes no llevan clave.
func (s *Store) VerifyAuditChain() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	previous := "genesis"
	for i, entry := range s.logs {
		if entry.PrevHash != previous || entry.Hash != hashLog(entry) {
			return fmt.Errorf("%w: entrada %d (id %s)", ErrAuditTampered, i, entry.ID)
		}
		previous = entry.Hash
	}
	return nil
}

func hashLog(entry models.AuditLog) string {
	data := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%d|%s",
		entry.ID, entry.Timestamp, entry.ActorID, entry.ActorRole, entry.Method,
		entry.Endpoint, entry.StatusResponse, entry.PrevHash)
	sum := sha256.Sum256([]byte(data))
	return hex.EncodeToString(sum[:])
}

func newID() (string, error) {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generando identificador: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

func now() string { return time.Now().UTC().Format(time.RFC3339) }

func (s *Store) load(name string, dst any) error {
	data, err := os.ReadFile(filepath.Join(s.dir, name))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dst)
}

// save escribe en un archivo temporal restrictivo y lo renombra de forma atómica
// sobre el destino, para que una escritura fallida no deje un JSON a medias.
func (s *Store) save(name string, src any) error {
	data, err := json.MarshalIndent(src, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(s.dir, name)
	temporary := path + ".tmp"
	if err := os.WriteFile(temporary, data, 0o600); err != nil {
		return err
	}
	if err := os.Rename(temporary, path); err != nil {
		_ = os.Remove(temporary)
		return err
	}
	return nil
}
