package service

import (
	"odomosml/internal/customer/model"
	"odomosml/internal/customer/repository"
)

// CustomerService definieert de interface voor klant-gerelateerde operaties
type CustomerService interface {
	GetAllCustomers(filter model.CustomerFilter) ([]model.Customer, error)
	GetCustomerByID(id string) (*model.Customer, error)
	CreateCustomer(customer model.Customer) (*model.Customer, error)
	UpdateCustomer(id string, customer model.Customer) (*model.Customer, error)
	PartialUpdateCustomer(id string, updates map[string]interface{}) (*model.Customer, error)
	DeleteCustomer(id string) error
}

// customerService is de concrete implementatie van CustomerService
type customerService struct {
	repo repository.CustomerRepository
}

// NewCustomerService creëert een nieuwe instantie van customerService
func NewCustomerService(repo repository.CustomerRepository) CustomerService {
	return &customerService{repo: repo}
}

// GetAllCustomers haalt alle klanten op met de gegeven filters
func (s *customerService) GetAllCustomers(filter model.CustomerFilter) ([]model.Customer, error) {
	return s.repo.FindAll(filter)
}

// GetCustomerByID haalt een specifieke klant op basis van ID
func (s *customerService) GetCustomerByID(id string) (*model.Customer, error) {
	return s.repo.FindByID(id)
}

// CreateCustomer maakt een nieuwe klant aan
func (s *customerService) CreateCustomer(customer model.Customer) (*model.Customer, error) {
	err := s.repo.Create(&customer)
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

// UpdateCustomer werkt een bestaande klant volledig bij
func (s *customerService) UpdateCustomer(id string, customer model.Customer) (*model.Customer, error) {
	existing, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// Behoud created_at van bestaande klant
	customer.ID = existing.ID
	customer.CreatedAt = existing.CreatedAt
	err = s.repo.Update(&customer)
	if err != nil {
		return nil, err
	}

	return &customer, nil
}

// PartialUpdateCustomer werkt een bestaande klant gedeeltelijk bij
func (s *customerService) PartialUpdateCustomer(id string, updates map[string]interface{}) (*model.Customer, error) {
	customer, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// Update alleen de velden die zijn meegegeven
	if name, ok := updates["name"].(string); ok {
		customer.Name = name
	}
	if email, ok := updates["email"].(string); ok {
		customer.Email = email
	}
	if phone, ok := updates["phone"].(string); ok {
		customer.Phone = phone
	}
	if address, ok := updates["address"].(string); ok {
		customer.Address = address
	}

	err = s.repo.Update(customer)
	if err != nil {
		return nil, err
	}

	return customer, nil
}

// DeleteCustomer verwijdert een klant
func (s *customerService) DeleteCustomer(id string) error {
	return s.repo.Delete(id)
}
