package service

import (
	"errors"
	"odomosml/internal/user/model"
	"odomosml/internal/user/repository"
)

type UserService interface {
	GetAllUsers() ([]model.UserResponse, error)
	GetUserByID(id string) (*model.UserResponse, error)
	CreateUser(user model.User) (*model.UserResponse, error)
	UpdateUser(id string, user model.User) (*model.UserResponse, error)
	DeleteUser(id string) error
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) GetAllUsers() ([]model.UserResponse, error) {
	users, err := s.repo.FindAll()
	if err != nil {
		return nil, err
	}

	userResponses := make([]model.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = user.ToResponse()
	}
	return userResponses, nil
}

func (s *userService) GetUserByID(id string) (*model.UserResponse, error) {
	user, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	response := user.ToResponse()
	return &response, nil
}

func (s *userService) CreateUser(user model.User) (*model.UserResponse, error) {
	// Check if email already exists
	existing, _ := s.repo.FindByEmail(user.Email)
	if existing != nil {
		return nil, errors.New("email is al in gebruik")
	}

	err := s.repo.Create(&user)
	if err != nil {
		return nil, err
	}

	response := user.ToResponse()
	return &response, nil
}

func (s *userService) UpdateUser(id string, user model.User) (*model.UserResponse, error) {
	existing, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// Update fields
	existing.Username = user.Username
	existing.Email = user.Email
	existing.Role = user.Role
	existing.Active = user.Active

	// Only update password if provided
	if user.Password != "" {
		existing.Password = user.Password
	}

	err = s.repo.Update(existing)
	if err != nil {
		return nil, err
	}

	response := existing.ToResponse()
	return &response, nil
}

func (s *userService) DeleteUser(id string) error {
	return s.repo.Delete(id)
}
