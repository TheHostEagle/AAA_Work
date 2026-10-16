// Package storage implements the project's JSON-backed persistence.
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

var (
	ErrNotFound       = errors.New("registro no encontrado")
	ErrEmailTaken     = errors.New("el correo ya está registrado")
	ErrStatusMismatch = errors.New("estado actual incompatible con la operación")
	ErrAuditTampered  = errors.New("cadena de auditoría alterada")
)

// Store guards in-memory records and their JSON files.
type Store struct {
	mu       sync.Mutex
	dir      string
	users    []models.User
	vehicles []models.Vehicle
	logs     []models.AuditLog
}

// New creates the data directory and loads persisted records.
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

// CreateUser assigns an ID and creation time, then persists the new user.
func (s *Store) CreateUser(user *models.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, existing := range s.users {
		if existing.Email == user.Email {
			return ErrEmailTaken
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
	return models.User{}, ErrNotFound
}

func (s *Store) GetUserByID(id string) (models.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, user := range s.users {
		if user.ID == id {
			return user, nil
		}
	}
	return models.User{}, ErrNotFound
}

// DeleteUser removes only the user record. The service layer decides how the
// user's vehicle records should be handled.
func (s *Store) DeleteUser(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, user := range s.users {
		if user.ID != id {
			continue
		}
		next := append([]models.User(nil), s.users[:i]...)
		next = append(next, s.users[i+1:]...)
		if err := s.save("users.json", next); err != nil {
			return err
		}
		s.users = next
		return nil
	}
	return ErrNotFound
}

// CreateVehicle assigns an ID and creation time. The service supplies its
// initial state and owner before calling this persistence method.
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
	return models.Vehicle{}, ErrNotFound
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

func (s *Store) ListVehiclesBySeller(sellerID string) []models.Vehicle {
	s.mu.Lock()
	defer s.mu.Unlock()
	vehicles := make([]models.Vehicle, 0)
	for _, vehicle := range s.vehicles {
		if vehicle.SellerID == sellerID {
			vehicles = append(vehicles, vehicle)
		}
	}
	return vehicles
}

// ChangeVehicleStatus checks the current state and persists the transition
// while holding one lock, so concurrent transitions cannot overwrite each other.
func (s *Store) ChangeVehicleStatus(id, from, to string) (models.Vehicle, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, vehicle := range s.vehicles {
		if vehicle.ID != id {
			continue
		}
		if vehicle.Status != from {
			return models.Vehicle{}, ErrStatusMismatch
		}
		next := append([]models.Vehicle(nil), s.vehicles...)
		next[i].Status = to
		if err := s.save("vehicles.json", next); err != nil {
			return models.Vehicle{}, err
		}
		s.vehicles = next
		return next[i], nil
	}
	return models.Vehicle{}, ErrNotFound
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
	return ErrNotFound
}

// SaveAuditLog completes the chain fields and appends the entry only after the
// new complete file has been persisted successfully.
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

// VerifyAuditChain detects broken links or entries whose stored contents no
// longer match their hash. It does not prevent an attacker from rebuilding the
// entire chain because the hashes are not keyed.
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

// save uses a restrictive temporary file and atomically renames it over the
// target so a failed write does not publish a partial JSON document.
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
