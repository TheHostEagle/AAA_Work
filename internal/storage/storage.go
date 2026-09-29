// implementa la persistencia nativa del proyecto
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
	ErrNotFound   = errors.New("registro no encontrado")
	ErrEmailTaken = errors.New("el correo ya está registrado")
)

// Store guarda usuarios, vehículos y logs de auditoría.
type Store struct {
	mu       sync.Mutex
	dir      string
	users    []models.User
	vehicles []models.Vehicle
	logs     []models.AuditLog
}

// New crea (si no existe) el directorio y carga los datos ya guardados.
func New(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
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

//USUARIOS

// CreateUser asigna ID y CreatedAt, y guarda el usuario.
// Devuelve ErrEmailTaken si el correo ya existe.
func (s *Store) CreateUser(u *models.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, x := range s.users {
		if x.Email == u.Email {
			return ErrEmailTaken
		}
	}
	u.ID = newID()
	u.CreatedAt = now()
	s.users = append(s.users, *u)
	return s.save("users.json", s.users)
}

// GetUserByEmail se usa en el login.
func (s *Store) GetUserByEmail(email string) (models.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, u := range s.users {
		if u.Email == email {
			return u, nil
		}
	}
	return models.User{}, ErrNotFound
}

// GetUserByID se usa para obtener el contacto del vendedor de un vehículo.
func (s *Store) GetUserByID(id string) (models.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, u := range s.users {
		if u.ID == id {
			return u, nil
		}
	}
	return models.User{}, ErrNotFound
}

// DeleteUser elimina un usuario por ID (HU-ADM-01).
func (s *Store) DeleteUser(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, u := range s.users {
		if u.ID == id {
			s.users = append(s.users[:i], s.users[i+1:]...)
			return s.save("users.json", s.users)
		}
	}
	return ErrNotFound
}

// VEHICULOS

// CreateVehicle asigna ID y CreatedAt, y lo deja en estado "pendiente".
func (s *Store) CreateVehicle(v *models.Vehicle) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	v.ID = newID()
	v.CreatedAt = now()
	v.Status = "pendiente"
	s.vehicles = append(s.vehicles, *v)
	return s.save("vehicles.json", s.vehicles)
}

// GetVehicle busca un vehículo por ID.
func (s *Store) GetVehicle(id string) (models.Vehicle, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, v := range s.vehicles {
		if v.ID == id {
			return v, nil
		}
	}
	return models.Vehicle{}, ErrNotFound
}

// ListVehiclesByStatus devuelve los vehículos con ese estado
// (el catálogo público usa "publicada").
func (s *Store) ListVehiclesByStatus(status string) []models.Vehicle {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := []models.Vehicle{}
	for _, v := range s.vehicles {
		if v.Status == status {
			out = append(out, v)
		}
	}
	return out
}

// UpdateVehicle reemplaza el vehículo que tenga el mismo ID
// (aprobar publicación, reportar venta).
func (s *Store) UpdateVehicle(v models.Vehicle) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.vehicles {
		if s.vehicles[i].ID == v.ID {
			s.vehicles[i] = v
			return s.save("vehicles.json", s.vehicles)
		}
	}
	return ErrNotFound
}

// DeleteVehicle elimina una publicación por ID (HU-ADM-01).
func (s *Store) DeleteVehicle(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, v := range s.vehicles {
		if v.ID == id {
			s.vehicles = append(s.vehicles[:i], s.vehicles[i+1:]...)
			return s.save("vehicles.json", s.vehicles)
		}
	}
	return ErrNotFound
}

// ACCOUTING

// SaveAuditLog completa ID, Timestamp, PrevHash y Hash, y agrega la entrada.
// Cada entrada queda encadenada a la anterior: si alguien modifica un log
// viejo, su hash deja de coincidir con el PrevHash de la siguiente.
func (s *Store) SaveAuditLog(l *models.AuditLog) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	l.ID = newID()
	l.Timestamp = now()
	l.PrevHash = "genesis"
	if n := len(s.logs); n > 0 {
		l.PrevHash = s.logs[n-1].Hash
	}
	l.Hash = hashLog(*l)
	s.logs = append(s.logs, *l)
	return s.save("audit.json", s.logs)
}

func hashLog(l models.AuditLog) string {
	data := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%d|%s",
		l.ID, l.Timestamp, l.ActorID, l.ActorRole, l.Method, l.Endpoint, l.StatusResponse, l.PrevHash)
	sum := sha256.Sum256([]byte(data))
	return hex.EncodeToString(sum[:])
}

//AUXILIARES

func newID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func now() string { return time.Now().UTC().Format(time.RFC3339) }

func (s *Store) load(name string, dst any) error {
	data, err := os.ReadFile(filepath.Join(s.dir, name))
	if errors.Is(err, os.ErrNotExist) {
		return nil // primera ejecución: arranca vacío
	}
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dst)
}

// save escribe en un archivo temporal y lo renombra, para no dejar
// un JSON a medias si el programa se cae escribiendo.
func (s *Store) save(name string, src any) error {
	data, err := json.MarshalIndent(src, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(s.dir, name)
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
