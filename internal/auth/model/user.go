package model

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Username  string    `json:"username" binding:"required"`
	Email     string    `json:"email" binding:"required,email"`
	Password  string    `json:"password,omitempty" binding:"required"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"` // seconds until expiration
	Username    string `json:"username"`
	Email       string `json:"email"`
	Role        string `json:"role"`
}

type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// Implementeer de jwt.Claims interface methodes
func (c *Claims) GetExpirationTime() (*jwt.NumericDate, error) {
	return c.ExpiresAt, nil
}

func (c *Claims) GetIssuedAt() (*jwt.NumericDate, error) {
	return c.IssuedAt, nil
}

func (c *Claims) GetNotBefore() (*jwt.NumericDate, error) {
	return c.NotBefore, nil
}

func (c *Claims) GetIssuer() (string, error) {
	return c.Issuer, nil
}

func (c *Claims) GetSubject() (string, error) {
	return c.Subject, nil
}

func (c *Claims) GetAudience() (jwt.ClaimStrings, error) {
	return c.Audience, nil
}
