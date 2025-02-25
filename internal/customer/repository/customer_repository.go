package repository

import (
	"errors"
	"odomosml/internal/customer/model"

	"gorm.io/gorm"
)

// CustomerRepository definieert de interface voor database operaties
type CustomerRepository interface {
	FindAll(filter model.CustomerFilter) ([]model.Customer, error)
	FindByID(id string) (*model.Customer, error)
	Create(customer *model.Customer) error
	Update(customer *model.Customer) error
	Delete(id string) error
	GetCustomerForAudit(id string) (map[string]interface{}, error)
}

type customerRepository struct {
	db *gorm.DB
}

// NewCustomerRepository maakt een nieuwe repository instantie
func NewCustomerRepository(db *gorm.DB) CustomerRepository {
	return &customerRepository{db: db}
}

// GetCustomerForAudit haalt klantgegevens op voor audit logging
func (r *customerRepository) GetCustomerForAudit(id string) (map[string]interface{}, error) {
	var customer model.Customer
	if err := r.db.First(&customer, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return map[string]interface{}{
		"name":    customer.Name,
		"email":   customer.Email,
		"phone":   customer.Phone,
		"address": customer.Address,
	}, nil
}

// FindAll haalt alle klanten op met de gegeven filters
func (r *customerRepository) FindAll(filter model.CustomerFilter) ([]model.Customer, error) {
	var customers []model.Customer
	query := r.db.Model(&model.Customer{})

	if filter.SearchTerm != "" {
		searchTerm := "%" + filter.SearchTerm + "%"
		query = query.Where("name ILIKE ? OR email ILIKE ?", searchTerm, searchTerm)
	}

	if filter.SortBy != "" {
		direction := "ASC"
		if filter.SortOrder == "desc" {
			direction = "DESC"
		}
		query = query.Order(filter.SortBy + " " + direction)
	}

	// Paginering toepassen
	offset := (filter.Page - 1) * filter.PageSize
	err := query.Offset(offset).Limit(filter.PageSize).Find(&customers).Error
	if err != nil {
		return nil, err
	}

	return customers, nil
}

// FindByID zoekt een klant op ID
func (r *customerRepository) FindByID(id string) (*model.Customer, error) {
	var customer model.Customer
	if err := r.db.First(&customer, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("klant niet gevonden")
		}
		return nil, err
	}
	return &customer, nil
}

// Create maakt een nieuwe klant aan
func (r *customerRepository) Create(customer *model.Customer) error {
	return r.db.Create(customer).Error
}

// Update werkt een bestaande klant bij
func (r *customerRepository) Update(customer *model.Customer) error {
	return r.db.Save(customer).Error
}

// Delete verwijdert een klant
func (r *customerRepository) Delete(id string) error {
	return r.db.Delete(&model.Customer{}, "id = ?", id).Error
}
