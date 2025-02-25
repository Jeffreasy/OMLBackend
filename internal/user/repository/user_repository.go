package repository

import (
	"errors"
	"odomosml/internal/user/model"

	"gorm.io/gorm"
)

// Constanten voor rollen (voor gebruik in middleware)
const (
	RoleAdmin = string(model.RoleAdmin)
	RoleUser  = string(model.RoleUser)
)

// UserRepository interface definieert de methodes voor gebruikersbeheer
type UserRepository interface {
	FindAll(filter model.UserFilter) ([]model.User, int64, error)
	FindByID(id string) (*model.User, error)
	FindByEmail(email string) (*model.User, error)
	Create(user *model.User) error
	Update(user *model.User) error
	Delete(id string) error
	DeleteUser(id string) (map[string]interface{}, error)
}

// userRepository implementeert de UserRepository interface
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository maakt een nieuwe UserRepository instantie
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

// FindAll haalt alle gebruikers op met filters
func (r *userRepository) FindAll(filter model.UserFilter) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	// Bouw query op
	query := r.db.Model(&model.User{})

	// Filters toepassen
	if filter.SearchTerm != "" {
		searchTerm := "%" + filter.SearchTerm + "%"
		query = query.Where("username ILIKE ? OR email ILIKE ?", searchTerm, searchTerm)
	}

	if filter.Role != "" {
		query = query.Where("role = ?", filter.Role)
	}

	if filter.Active != nil {
		query = query.Where("active = ?", *filter.Active)
	}

	// Tel totaal aantal records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Paginering toepassen
	if filter.Page > 0 && filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		query = query.Offset(offset).Limit(filter.PageSize)
	}

	// Voer query uit
	if err := query.Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// FindByID haalt een gebruiker op op basis van ID
func (r *userRepository) FindByID(id string) (*model.User, error) {
	var user model.User

	if err := r.db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("gebruiker niet gevonden")
		}
		return nil, err
	}

	return &user, nil
}

// FindByEmail haalt een gebruiker op op basis van email
func (r *userRepository) FindByEmail(email string) (*model.User, error) {
	var user model.User

	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("gebruiker niet gevonden")
		}
		return nil, err
	}

	return &user, nil
}

// Create maakt een nieuwe gebruiker aan
func (r *userRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

// Update werkt een bestaande gebruiker bij
func (r *userRepository) Update(user *model.User) error {
	return r.db.Save(user).Error
}

// Delete verwijdert een gebruiker
// DEPRECATED: Gebruik DeleteUser in plaats hiervan
func (r *userRepository) Delete(id string) error {
	return r.db.Delete(&model.User{}, "id = ?", id).Error
}

// DeleteUser verwijdert een gebruiker en retourneert de gebruikersdata voor audit logging
func (r *userRepository) DeleteUser(id string) (map[string]interface{}, error) {
	// Haal gebruiker op voor audit logging
	user, err := r.FindByID(id)
	if err != nil {
		return nil, err
	}

	// Converteer naar map voor audit logging
	userData := map[string]interface{}{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
		"role":     user.Role,
	}

	// Verwijder gebruiker
	if err := r.db.Delete(&model.User{}, "id = ?", id).Error; err != nil {
		return nil, err
	}

	return userData, nil
}
