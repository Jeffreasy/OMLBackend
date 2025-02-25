package service

import (
	"errors"
	"odomosml/config"
	"odomosml/internal/auth/model"
	userModel "odomosml/internal/user/model"
	"odomosml/internal/user/repository"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AuthService interface {
	Login(email, password string) (*model.TokenResponse, error)
	Register(req model.RegisterRequest) (*model.TokenResponse, error)
	ValidateToken(tokenString string) (*model.Claims, error)
	RefreshToken(claims *model.Claims) (*model.TokenResponse, error)
}

type authService struct {
	userRepo repository.UserRepository
	config   *config.Config
}

func NewAuthService(userRepo repository.UserRepository, cfg *config.Config) AuthService {
	return &authService{
		userRepo: userRepo,
		config:   cfg,
	}
}

func (s *authService) Login(email, password string) (*model.TokenResponse, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, errors.New("ongeldige inloggegevens")
	}

	if err := user.ComparePassword(password); err != nil {
		return nil, errors.New("ongeldige inloggegevens")
	}

	if !user.Active {
		return nil, errors.New("account is gedeactiveerd")
	}

	return s.generateToken(user)
}

func (s *authService) Register(req model.RegisterRequest) (*model.TokenResponse, error) {
	// Check if email already exists
	if existing, _ := s.userRepo.FindByEmail(req.Email); existing != nil {
		return nil, errors.New("email is al in gebruik")
	}

	// Create new user
	user := &userModel.User{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		Role:     userModel.RoleUser, // Default role
		Active:   true,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	return s.generateToken(user)
}

func (s *authService) ValidateToken(tokenString string) (*model.Claims, error) {
	claims := &model.Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("ongeldig token type")
		}
		return []byte(s.config.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("ongeldig token")
	}

	return claims, nil
}

func (s *authService) generateToken(user *userModel.User) (*model.TokenResponse, error) {
	// Token expiration time (24 hours)
	expirationTime := time.Now().Add(24 * time.Hour)

	claims := &model.Claims{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     string(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		return nil, err
	}

	return &model.TokenResponse{
		AccessToken: tokenString,
		TokenType:   "Bearer",
		ExpiresIn:   int64(time.Until(expirationTime).Seconds()),
		Username:    user.Username,
		Email:       user.Email,
		Role:        string(user.Role),
	}, nil
}

func (s *authService) RefreshToken(claims *model.Claims) (*model.TokenResponse, error) {
	// Haal de gebruiker op om te verifiëren dat deze nog bestaat en actief is
	user, err := s.userRepo.FindByEmail(claims.Email)
	if err != nil {
		return nil, errors.New("gebruiker niet gevonden")
	}

	if !user.Active {
		return nil, errors.New("account is gedeactiveerd")
	}

	// Genereer een nieuwe token met vernieuwde expiratie
	expirationTime := time.Now().Add(24 * time.Hour)
	newClaims := &model.Claims{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     string(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims)
	tokenString, err := token.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		return nil, err
	}

	return &model.TokenResponse{
		AccessToken: tokenString,
		TokenType:   "Bearer",
		ExpiresIn:   int64(time.Until(expirationTime).Seconds()),
		Username:    user.Username,
		Email:       user.Email,
		Role:        string(user.Role),
	}, nil
}
