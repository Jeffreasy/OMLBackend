package repository

import (
	"errors"
	"odomosml/internal/user/model"

	"gorm.io/gorm"
)

type UserRepository interface {
	FindAll() ([]model.User, error)
	FindByID(id string) (*model.User, error)
	FindByEmail(email string) (*model.User, error)
	Create(user *model.User) error
	Update(user *model.User) error
	Delete(id string) error
	GetUserForAudit(id string) (map[string]interface{}, error)
	DeleteUser(id string) (map[string]interface{}, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindAll() ([]model.User, error) {
	var users []model.User
	if err := r.db.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *userRepository) FindByID(id string) (*model.User, error) {
	var user model.User
	if err := r.db.First(&user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("gebruiker niet gevonden")
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByEmail(email string) (*model.User, error) {
	var user model.User
	if err := r.db.First(&user, "email = ?", email).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("gebruiker niet gevonden")
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) Update(user *model.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) Delete(id string) error {
	return r.db.Delete(&model.User{}, "id = ?", id).Error
}

// GetUserForAudit haalt gebruikersgegevens op voor audit logging
func (r *userRepository) GetUserForAudit(id string) (map[string]interface{}, error) {
	var user model.User
	if err := r.db.First(&user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return map[string]interface{}{
		"username": user.Username,
		"email":    user.Email,
		"role":     user.Role,
		"active":   user.Active,
	}, nil
}

func (r *userRepository) DeleteUser(id string) (map[string]interface{}, error) {
	// Eerst de gebruiker ophalen voor audit logging
	var user model.User
	if err := r.db.Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}

	// Gebruikersgegevens opslaan voor audit
	userData := map[string]interface{}{
		"username": user.Username,
		"email":    user.Email,
		"role":     user.Role,
	}

	// Dan de gebruiker verwijderen
	if err := r.db.Delete(&model.User{}, id).Error; err != nil {
		return nil, err
	}

	return userData, nil
}
