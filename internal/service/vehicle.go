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

// VehicleService contiene las reglas de catálogo y publicación (HU-VIS-01, HU-VEN-01).
type VehicleService struct {
	store vehicleStore
}

func NewVehicleService(store vehicleStore) *VehicleService {
	return &VehicleService{store: store}
}

// PublicCatalog devuelve las publicaciones publicadas con el contacto del
// vendedor (HU-VIS-01).
func (s *VehicleService) PublicCatalog() ([]models.PublicVehicle, error) {
	vehicles := s.store.ListVehiclesByStatus(models.StatusPublished)
	listings := make([]models.PublicVehicle, 0, len(vehicles))
	for _, vehicle := range vehicles {
		seller, err := s.store.GetUserByID(vehicle.SellerID)
		if err != nil {
			return nil, fmt.Errorf("obteniendo contacto del vendedor %s: %w", vehicle.SellerID, err)
		}
		listings = append(listings, models.PublicVehicle{
			Vehicle: vehicle,
			Seller:  models.SellerContact{Name: seller.Name, Email: seller.Email},
		})
	}
	return listings, nil
}

// CreateForSeller registra la publicación en estado pendiente de aprobación,
// asociada al vendedor autenticado (HU-VEN-01).
func (s *VehicleService) CreateForSeller(sellerID string, vehicle models.Vehicle) (models.Vehicle, error) {
	vehicle.ID = ""
	vehicle.SellerID = sellerID
	vehicle.Status = models.StatusPending
	if err := s.store.CreateVehicle(&vehicle); err != nil {
		return models.Vehicle{}, err
	}
	return vehicle, nil
}

// ReportSold permite al dueño marcar como vendida una publicación publicada.
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
