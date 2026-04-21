package models

import (
	"encoding/json"
	"time"
)

type App struct {
	ID          string          `json:"id"`
	UserID      string          `json:"user_id"`
	Slug        string          `json:"slug"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	HTMLContent string          `json:"html_content"`
	Category    json.RawMessage `json:"category"`
	IsHidden    bool            `json:"is_hidden"`
	IsPublic    bool            `json:"is_public"`
	Approved    bool            `json:"approved"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	// Joined from users table
	OwnerPYCCode string `json:"owner_pyccode,omitempty"`
}
