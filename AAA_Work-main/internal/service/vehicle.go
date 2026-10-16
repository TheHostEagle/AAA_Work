package service

import (
	"fmt"

	"automarket/internal/models"
)

type vehicleStore interface {
	CreateVehicle(*models.Vehicle) error
	GetVehicle(string) (models.Vehicle, error)
	ListVehiclesByStatus(string) []models.Vehicle
	GetUserByID(string) (models.User, error)
	ChangeVehicleStatus(id, from, to string) (models.Vehicle, error)
}

type VehicleService struct {
	store vehicleStore
}

func NewVehicleService(store vehicleStore) *VehicleService {
	return &VehicleService{store: store}
}

func (s *VehicleService) PublicCatalog() ([]models.PublicVehicle, error) {
	vehicles := s.store.ListVehiclesByStatus(models.StatusPublished)
	listings := make([]models.PublicVehicle, 0, len(vehicles))
	for _, vehicle := range vehicles {
		seller, err := s.store.GetUserByID(vehicle.SellerID)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrCatalogUnavailable, err)
		}
		listings = append(listings, models.PublicVehicle{
			Vehicle: vehicle,
			Seller:  models.SellerContact{Name: seller.Name, Email: seller.Email},
		})
	}
	return listings, nil
}

func (s *VehicleService) CreateForSeller(sellerID string, vehicle models.Vehicle) (models.Vehicle, error) {
	vehicle.ID = ""
	vehicle.SellerID = sellerID
	vehicle.Status = models.StatusPending
	if err := s.store.CreateVehicle(&vehicle); err != nil {
		return models.Vehicle{}, err
	}
	return vehicle, nil
}

func (s *VehicleService) ReportSold(sellerID, vehicleID string) (models.Vehicle, error) {
	vehicle, err := s.store.GetVehicle(vehicleID)
	if err != nil {
		return models.Vehicle{}, err
	}
	if vehicle.SellerID != sellerID {
		return models.Vehicle{}, ErrForbidden
	}
	return s.store.ChangeVehicleStatus(vehicleID, models.StatusPublished, models.StatusSold)
}
