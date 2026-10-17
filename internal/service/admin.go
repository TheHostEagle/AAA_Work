package service

import "automarket/internal/models"

type adminStore interface {
	DeleteVehicle(id string) error
	DeleteUserWithVehicles(id string) error
	ChangeVehicleStatus(id, from, to string) (models.Vehicle, error)
}

// AdminService contiene las reglas de moderación de HU-ADM-01.
type AdminService struct {
	store adminStore
}

func NewAdminService(store adminStore) *AdminService {
	return &AdminService{store: store}
}

// ApproveVehicle pasa una publicación de pendiente a publicada.
func (s *AdminService) ApproveVehicle(vehicleID string) (models.Vehicle, error) {
	return s.store.ChangeVehicleStatus(vehicleID, models.StatusPending, models.StatusPublished)
}

func (s *AdminService) DeleteVehicle(vehicleID string) error {
	return s.store.DeleteVehicle(vehicleID)
}

// DeleteUser elimina la cuenta junto con sus publicaciones.
func (s *AdminService) DeleteUser(userID string) error {
	return s.store.DeleteUserWithVehicles(userID)
}
