package service

import (
	"errors"
	"odomosml/internal/user/model"
	"odomosml/internal/user/repository"
)

type UserService interface {
	GetAllUsers() ([]model.User, error)
	GetUserByID(id string) (*model.User, error)
	CreateUser(user model.User) (*model.User, error)
	UpdateUser(id string, user model.User) (*model.User, error)
	DeleteUser(id string) (map[string]interface{}, error)
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) GetAllUsers() ([]model.User, error) {
	return s.repo.FindAll()
}

func (s *userService) GetUserByID(id string) (*model.User, error) {
	return s.repo.FindByID(id)
}

func (s *userService) CreateUser(user model.User) (*model.User, error) {
	// Check if email already exists
	existing, _ := s.repo.FindByEmail(user.Email)
	if existing != nil {
		return nil, errors.New("email is al in gebruik")
	}

	if err := s.repo.Create(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *userService) UpdateUser(id string, user model.User) (*model.User, error) {
	existingUser, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// Update fields
	existingUser.Username = user.Username
	existingUser.Email = user.Email
	existingUser.Role = user.Role
	existingUser.Active = user.Active

	// Update password only if provided
	if user.Password != "" {
		existingUser.Password = user.Password
	}

	if err := s.repo.Update(existingUser); err != nil {
		return nil, err
	}
	return existingUser, nil
}

func (s *userService) DeleteUser(id string) (map[string]interface{}, error) {
	return s.repo.DeleteUser(id)
}
