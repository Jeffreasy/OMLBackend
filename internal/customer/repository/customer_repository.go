package repository

import (
	"errors"
	"odomosml/internal/customer/model"
	"strconv"

	"gorm.io/gorm"
)

// CustomerRepository definieert de interface voor customer repository
type CustomerRepository interface {
	FindAll(filter model.CustomerFilter) ([]model.Customer, int64, error)
	FindByID(id string) (*model.Customer, error)
	Create(customer *model.Customer) (*model.Customer, error)
	Update(customer *model.Customer) (*model.Customer, error)
	PartialUpdate(id uint, updates map[string]interface{}) (*model.Customer, error)
	Delete(id string) error
}

// customerRepository implementeert de CustomerRepository interface
type customerRepository struct {
	db *gorm.DB
}

// NewCustomerRepository maakt een nieuwe CustomerRepository instantie
func NewCustomerRepository(db *gorm.DB) CustomerRepository {
	return &customerRepository{
		db: db,
	}
}

// FindAll haalt alle klanten op met filters
func (r *customerRepository) FindAll(filter model.CustomerFilter) ([]model.Customer, int64, error) {
	var customers []model.Customer
	var total int64

	// Bouw query
	query := r.db.Model(&model.Customer{})

	// Filters toepassen
	if filter.SearchTerm != "" {
		searchTerm := "%" + filter.SearchTerm + "%"
		query = query.Where("name ILIKE ? OR email ILIKE ?", searchTerm, searchTerm)
	}

	// Tel totaal aantal records (voor paginering)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Paginering toepassen
	offset := (filter.Page - 1) * filter.PageSize
	query = query.Offset(offset).Limit(filter.PageSize)

	// Sorteer op naam
	query = query.Order("name ASC")

	// Voer query uit
	if err := query.Find(&customers).Error; err != nil {
		return nil, 0, err
	}

	return customers, total, nil
}

// FindByID haalt een klant op op basis van ID
func (r *customerRepository) FindByID(id string) (*model.Customer, error) {
	var customer model.Customer

	// Converteer string ID naar uint
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return nil, errors.New("ongeldig ID formaat")
	}

	// Zoek klant
	if err := r.db.First(&customer, idInt).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("klant niet gevonden")
		}
		return nil, err
	}

	return &customer, nil
}

// Create maakt een nieuwe klant aan
func (r *customerRepository) Create(customer *model.Customer) (*model.Customer, error) {
	if err := r.db.Create(customer).Error; err != nil {
		return nil, err
	}
	return customer, nil
}

// Update werkt een bestaande klant bij
func (r *customerRepository) Update(customer *model.Customer) (*model.Customer, error) {
	if err := r.db.Save(customer).Error; err != nil {
		return nil, err
	}
	return customer, nil
}

// PartialUpdate werkt een deel van een bestaande klant bij
func (r *customerRepository) PartialUpdate(id uint, updates map[string]interface{}) (*model.Customer, error) {
	// Update klant
	if err := r.db.Model(&model.Customer{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return nil, err
	}

	// Haal bijgewerkte klant op
	var customer model.Customer
	if err := r.db.First(&customer, id).Error; err != nil {
		return nil, err
	}

	return &customer, nil
}

// Delete verwijdert een klant
func (r *customerRepository) Delete(id string) error {
	// Converteer string ID naar uint
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return errors.New("ongeldig ID formaat")
	}

	// Verwijder klant
	if err := r.db.Delete(&model.Customer{}, idInt).Error; err != nil {
		return err
	}

	return nil
}
