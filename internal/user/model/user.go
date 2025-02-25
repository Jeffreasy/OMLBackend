package model

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Role definieert de gebruikersrol
// @Description Gebruikersrol (ADMIN of USER)
type Role string

// Beschikbare rollen
const (
	RoleAdmin Role = "ADMIN"
	RoleUser  Role = "USER"
)

// User represents a user in the system
// @Description Een gebruiker in het systeem
type User struct {
	ID        uint      `json:"id" gorm:"primaryKey" example:"1" swaggertype:"integer"`
	Username  string    `json:"username" gorm:"size:50;not null;unique" example:"johndoe" swaggertype:"string"`
	Email     string    `json:"email" gorm:"size:100;not null;unique" example:"john@example.com" swaggertype:"string"`
	Password  string    `json:"password,omitempty" gorm:"size:255;not null" example:"password123" swaggertype:"string"`
	Role      Role      `json:"role" gorm:"size:20;not null;default:'USER'" example:"USER" swaggertype:"string"`
	Active    bool      `json:"active" gorm:"default:true" example:"true" swaggertype:"boolean"`
	CreatedAt time.Time `json:"created_at" example:"2024-02-25T20:30:00Z" swaggertype:"string" format:"date-time"`
	UpdatedAt time.Time `json:"updated_at" example:"2024-02-25T20:30:00Z" swaggertype:"string" format:"date-time"`
}

// BeforeSave wordt aangeroepen voordat een gebruiker wordt opgeslagen
// Dit zorgt ervoor dat wachtwoorden altijd worden gehasht
func (u *User) BeforeSave(tx *gorm.DB) error {
	// Als het wachtwoord leeg is, doe niets (bijv. bij updates waar wachtwoord niet wordt gewijzigd)
	if u.Password == "" {
		return nil
	}

	// Controleer of het wachtwoord al gehasht is
	// bcrypt hashes beginnen met $2a$, $2b$ of $2y$
	if len(u.Password) >= 4 && (u.Password[:4] == "$2a$" || u.Password[:4] == "$2b$" || u.Password[:4] == "$2y$") {
		return nil // Wachtwoord is al gehasht
	}

	// Hash het wachtwoord
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("fout bij het hashen van wachtwoord")
	}

	u.Password = string(hashedPassword)
	return nil
}

// ComparePassword vergelijkt een plaintext wachtwoord met het gehashte wachtwoord
func (u *User) ComparePassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
}

// TableName specificeert de tabelnaam voor GORM
func (User) TableName() string {
	return "users"
}

// UserResponse is de response struct voor User data
// @Description Response object voor gebruikersgegevens
type UserResponse struct {
	ID        uint      `json:"id" example:"1" swaggertype:"integer"`
	Username  string    `json:"username" example:"johndoe" swaggertype:"string"`
	Email     string    `json:"email" example:"john@example.com" swaggertype:"string"`
	Role      Role      `json:"role" example:"USER" swaggertype:"string"`
	Active    bool      `json:"active" example:"true" swaggertype:"boolean"`
	CreatedAt time.Time `json:"created_at" example:"2024-02-25T20:30:00Z" swaggertype:"string" format:"date-time"`
	UpdatedAt time.Time `json:"updated_at" example:"2024-02-25T20:30:00Z" swaggertype:"string" format:"date-time"`
}

// ToResponse converteert een User naar een UserResponse (zonder wachtwoord)
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		Role:      u.Role,
		Active:    u.Active,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// UserFilter definieert filters voor het ophalen van gebruikers
// @Description Filter opties voor gebruikerslijsten
type UserFilter struct {
	SearchTerm string `json:"search_term" form:"search_term" example:"john" swaggertype:"string"`
	Role       Role   `json:"role" form:"role" example:"USER" swaggertype:"string"`
	Active     *bool  `json:"active" form:"active" example:"true" swaggertype:"boolean"`
	Page       int    `json:"page" form:"page" example:"1" swaggertype:"integer"`
	PageSize   int    `json:"page_size" form:"page_size" example:"10" swaggertype:"integer"`
}
