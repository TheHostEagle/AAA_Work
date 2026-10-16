package service

import (
	"automarket/internal/models"
)

type adminStore interface {
	GetUserByID(string) (models.User, error)
	ListVehiclesBySeller(string) []models.Vehicle
	DeleteVehicle(string) error
	DeleteUser(string) error
	ChangeVehicleStatus(id, from, to string) (models.Vehicle, error)
}

type AdminService struct {
	store adminStore
}

func NewAdminService(store adminStore) *AdminService {
	return &AdminService{store: store}
}

func (s *AdminService) ApproveVehicle(vehicleID string) (models.Vehicle, error) {
	return s.store.ChangeVehicleStatus(vehicleID, models.StatusPending, models.StatusPublished)
}

func (s *AdminService) DeleteVehicle(vehicleID string) error {
	return s.store.DeleteVehicle(vehicleID)
}

func (s *AdminService) DeleteUser(userID string) error {
	if _, err := s.store.GetUserByID(userID); err != nil {
		return err
	}
	for _, vehicle := range s.store.ListVehiclesBySeller(userID) {
		if err := s.store.DeleteVehicle(vehicle.ID); err != nil {
			return err
		}
	}
	return s.store.DeleteUser(userID)
}
