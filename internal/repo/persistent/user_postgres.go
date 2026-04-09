package persistent

import (
	"context"
	"fmt"
	"time"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/pkg/postgres"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// UserRepo -.
type UserRepo struct {
	*postgres.Postgres
}

// NewUserRepo -.
func NewUserRepo(pg *postgres.Postgres) *UserRepo {
	return &UserRepo{pg}
}

// CreateUser creates a new user
func (r *UserRepo) CreateUser(ctx context.Context, user entity.User) error {
	query := `
		INSERT INTO users(id, username, email, password, role, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("UserRepo - CreateUser - bcrypt.GenerateFromPassword: %w", err)
	}

	user.ID = uuid.New().String()
	now := time.Now()

	_, err = r.Pool.Exec(ctx, query,
		user.ID,
		user.Username,
		user.Email,
		string(hashedPassword),
		user.Role,
		now,
		now,
	)
	if err != nil {
		return fmt.Errorf("UserRepo - CreateUser - r.Pool.Exec: %w", err)
	}

	return nil
}

// GetUserByID retrieves a user by ID
func (r *UserRepo) GetUserByID(ctx context.Context, id string) (*entity.User, error) {
	query := `SELECT id, username, email, password, role, created_at, updated_at 
	          FROM users WHERE id = $1`

	var user entity.User
	err := r.Pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("UserRepo - GetUserByID - r.Pool.QueryRow: %w", err)
	}

	return &user, nil
}

// GetUserByUsername retrieves a user by username
func (r *UserRepo) GetUserByUsername(ctx context.Context, username string) (*entity.User, error) {
	query := `SELECT id, username, email, password, role, created_at, updated_at 
	          FROM users WHERE username = $1`

	var user entity.User
	err := r.Pool.QueryRow(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("UserRepo - GetUserByUsername - r.Pool.QueryRow: %w", err)
	}

	return &user, nil
}

// GetUserByEmail retrieves a user by email
func (r *UserRepo) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	query := `SELECT id, username, email, password, role, created_at, updated_at 
	          FROM users WHERE email = $1`

	var user entity.User
	err := r.Pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("UserRepo - GetUserByEmail - r.Pool.QueryRow: %w", err)
	}

	return &user, nil
}

// UpdateUser updates a user's information
func (r *UserRepo) UpdateUser(ctx context.Context, user entity.User) error {
	query := `UPDATE users SET username=$1, email=$2, role=$3, updated_at=$4 WHERE id=$5`

	_, err := r.Pool.Exec(ctx, query,
		user.Username,
		user.Email,
		user.Role,
		time.Now(),
		user.ID,
	)
	if err != nil {
		return fmt.Errorf("UserRepo - UpdateUser - r.Pool.Exec: %w", err)
	}

	return nil
}

// DeleteUser deletes a user
func (r *UserRepo) DeleteUser(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = $1`

	_, err := r.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("UserRepo - DeleteUser - r.Pool.Exec: %w", err)
	}

	return nil
}

// VerifyPassword verifies if the provided password matches the user's hash
func (r *UserRepo) VerifyPassword(ctx context.Context, username, password string) (bool, error) {
	user, err := r.GetUserByUsername(ctx, username)
	if err != nil {
		return false, fmt.Errorf("UserRepo - VerifyPassword - r.GetUserByUsername: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return false, nil // Return false, not an error for invalid password
	}

	return true, nil
}

// UpdatePassword updates a user's password
func (r *UserRepo) UpdatePassword(ctx context.Context, userID, newPassword string) error {
	query := `UPDATE users SET password=$1, updated_at=$2 WHERE id=$3`

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("UserRepo - UpdatePassword - bcrypt.GenerateFromPassword: %w", err)
	}

	_, err = r.Pool.Exec(ctx, query,
		string(hashedPassword),
		time.Now(),
		userID,
	)
	if err != nil {
		return fmt.Errorf("UserRepo - UpdatePassword - r.Pool.Exec: %w", err)
	}

	return nil
}
