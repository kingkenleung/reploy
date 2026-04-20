package models

import "time"

type App struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Slug        string    `json:"slug"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	HTMLContent string    `json:"html_content"`
	IsHidden    bool      `json:"is_hidden"`
	IsPublic    bool      `json:"is_public"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	// Joined from users table
	OwnerPYCCode string `json:"owner_pyccode,omitempty"`
}
