package service

import (
	"errors"
	"odomosml/internal/user/model"
	"odomosml/internal/user/repository"
	"strconv"
)

// UserService interface definieert de methodes voor gebruikersbeheer
type UserService interface {
	GetAllUsers(filter model.UserFilter) ([]model.User, int64, error)
	GetUserByID(id string) (*model.User, error)
	CreateUser(user *model.User) (*model.User, error)
	UpdateUser(user *model.User) (*model.User, error)
	DeleteUser(id string) (map[string]interface{}, error)
}

// userService implementeert de UserService interface
type userService struct {
	repo repository.UserRepository
}

// NewUserService maakt een nieuwe UserService instantie
func NewUserService(repo repository.UserRepository) UserService {
	return &userService{
		repo: repo,
	}
}

// GetAllUsers haalt alle gebruikers op met filters
func (s *userService) GetAllUsers(filter model.UserFilter) ([]model.User, int64, error) {
	// Valideer paginering
	if filter.Page < 1 {
		filter.Page = 1
	}

	if filter.PageSize < 1 || filter.PageSize > 100 {
		filter.PageSize = 10 // Default page size
	}

	return s.repo.FindAll(filter)
}

// GetUserByID haalt een gebruiker op op basis van ID
func (s *userService) GetUserByID(id string) (*model.User, error) {
	return s.repo.FindByID(id)
}

// CreateUser maakt een nieuwe gebruiker aan
func (s *userService) CreateUser(user *model.User) (*model.User, error) {
	// Valideer gebruiker
	if user.Username == "" {
		return nil, errors.New("gebruikersnaam is verplicht")
	}

	if user.Email == "" {
		return nil, errors.New("email is verplicht")
	}

	if user.Password == "" {
		return nil, errors.New("wachtwoord is verplicht")
	}

	// Controleer of email al bestaat
	if existing, _ := s.repo.FindByEmail(user.Email); existing != nil {
		return nil, errors.New("email is al in gebruik")
	}

	// Maak gebruiker aan
	if err := s.repo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

// UpdateUser werkt een bestaande gebruiker bij
func (s *userService) UpdateUser(user *model.User) (*model.User, error) {
	// Controleer of gebruiker bestaat
	existing, err := s.repo.FindByID(strconv.FormatUint(uint64(user.ID), 10))
	if err != nil {
		return nil, err
	}

	// Controleer of email al in gebruik is door een andere gebruiker
	if user.Email != existing.Email {
		if existingWithEmail, _ := s.repo.FindByEmail(user.Email); existingWithEmail != nil {
			return nil, errors.New("email is al in gebruik")
		}
	}

	// Werk gebruiker bij
	if err := s.repo.Update(user); err != nil {
		return nil, err
	}

	return user, nil
}

// DeleteUser verwijdert een gebruiker
func (s *userService) DeleteUser(id string) (map[string]interface{}, error) {
	// Haal gebruiker op voor audit logging
	user, err := s.repo.FindByID(id)
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
	if err := s.repo.Delete(id); err != nil {
		return nil, err
	}

	return userData, nil
}
