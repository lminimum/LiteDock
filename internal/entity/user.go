package entity

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID        string    `json:"id" db:"id"`
	Username  string    `json:"username" db:"username"`
	Email     string    `json:"email" db:"email"`
	Password  string    `json:"-" db:"password"` // never expose password in JSON
	Role      string    `json:"role" db:"role"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// UserAuth represents authentication data
type UserAuth struct {
	UserID   string `json:"user_id"`
	Token    string `json:"token"`
	ExpireAt int64  `json:"expire_at"`
}
