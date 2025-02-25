package service

import (
	"errors"
	"odomosml/internal/customer/model"
	"odomosml/internal/customer/repository"
)

// CustomerService definieert de interface voor customer service
type CustomerService interface {
	GetAllCustomers(filter model.CustomerFilter) ([]model.Customer, int64, error)
	GetCustomerByID(id string) (*model.Customer, error)
	CreateCustomer(customer *model.Customer) (*model.Customer, error)
	UpdateCustomer(customer *model.Customer) (*model.Customer, error)
	PartialUpdateCustomer(id string, updates map[string]interface{}) (*model.Customer, error)
	DeleteCustomer(id string) (map[string]interface{}, error)
}

// customerService implementeert de CustomerService interface
type customerService struct {
	repo repository.CustomerRepository
}

// NewCustomerService maakt een nieuwe CustomerService instantie
func NewCustomerService(repo repository.CustomerRepository) CustomerService {
	return &customerService{
		repo: repo,
	}
}

// GetAllCustomers haalt alle klanten op met filters
func (s *customerService) GetAllCustomers(filter model.CustomerFilter) ([]model.Customer, int64, error) {
	return s.repo.FindAll(filter)
}

// GetCustomerByID haalt een klant op op basis van ID
func (s *customerService) GetCustomerByID(id string) (*model.Customer, error) {
	return s.repo.FindByID(id)
}

// CreateCustomer maakt een nieuwe klant aan
func (s *customerService) CreateCustomer(customer *model.Customer) (*model.Customer, error) {
	// Validatie
	if customer.Name == "" {
		return nil, errors.New("naam is verplicht")
	}

	return s.repo.Create(customer)
}

// UpdateCustomer werkt een bestaande klant bij
func (s *customerService) UpdateCustomer(customer *model.Customer) (*model.Customer, error) {
	// Validatie
	if customer.ID == 0 {
		return nil, errors.New("klant ID is verplicht")
	}

	if customer.Name == "" {
		return nil, errors.New("naam is verplicht")
	}

	return s.repo.Update(customer)
}

// PartialUpdateCustomer werkt een deel van een bestaande klant bij
func (s *customerService) PartialUpdateCustomer(id string, updates map[string]interface{}) (*model.Customer, error) {
	// Validatie
	if id == "" {
		return nil, errors.New("klant ID is verplicht")
	}

	// Controleer of klant bestaat
	customer, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// Update velden
	return s.repo.PartialUpdate(customer.ID, updates)
}

// DeleteCustomer verwijdert een klant
func (s *customerService) DeleteCustomer(id string) (map[string]interface{}, error) {
	// Haal klant op voor audit logging
	customer, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// Converteer naar map voor audit logging
	customerData := map[string]interface{}{
		"id":    customer.ID,
		"name":  customer.Name,
		"email": customer.Email,
	}

	// Verwijder klant
	if err := s.repo.Delete(id); err != nil {
		return nil, err
	}

	return customerData, nil
}
