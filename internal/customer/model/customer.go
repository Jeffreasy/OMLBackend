package model

import "time"

type Customer struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" binding:"required"`
	Email     string    `json:"email" binding:"required,email"`
	Phone     string    `json:"phone"`
	Address   string    `json:"address"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CustomerFilter struct {
	SearchTerm string
	Page       int
	PageSize   int
	SortBy     string
	SortOrder  string
}
