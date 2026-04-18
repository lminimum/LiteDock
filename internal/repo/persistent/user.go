package persistent

import (
	"context"
	"fmt"
	"time"

	"github.com/lminimum/LiteDock/internal/entity"
	"github.com/lminimum/LiteDock/pkg/database"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// UserRepo -.
type UserRepo struct {
	db database.DB
}

// NewUserRepo -.
func NewUserRepo(db database.DB) *UserRepo {
	return &UserRepo{db: db}
}

// scanRow is a helper to scan a row from QueryRow
func scanRow(row interface{}, dest ...interface{}) error {
	if row == nil {
		return fmt.Errorf("scanRow: nil row")
	}
	scanner, ok := row.(interface{ Scan(...interface{}) error })
	if !ok {
		return fmt.Errorf("scanRow: row does not implement Scanner interface")
	}

	timeIndices := make(map[int]*time.Time)
	tempDest := make([]interface{}, len(dest))
	for i, d := range dest {
		if t, ok := d.(*time.Time); ok {
			timeIndices[i] = t
			var bs []byte
			tempDest[i] = &bs
		} else {
			tempDest[i] = d
		}
	}

	if err := scanner.Scan(tempDest...); err != nil {
		return err
	}

	for i, t := range timeIndices {
		bs, ok := tempDest[i].(*[]byte)
		if ok && bs != nil && len(*bs) > 0 {
			parsed, err := time.Parse("2006-01-02 15:04:05", string(*bs))
			if err != nil {
				parsed, err = time.Parse(time.RFC3339, string(*bs))
			}
			if err == nil {
				*t = parsed
			}
		}
	}

	return nil
}

// CreateUser creates a new user
func (r *UserRepo) CreateUser(ctx context.Context, user entity.User) error {
	query := `
		INSERT INTO users(id, username, email, password, role, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("UserRepo - CreateUser - bcrypt.GenerateFromPassword: %w", err)
	}

	user.ID = uuid.New().String()
	now := time.Now()

	err = r.db.Exec(ctx, query,
		user.ID,
		user.Username,
		user.Email,
		string(hashedPassword),
		user.Role,
		now,
		now,
	)
	if err != nil {
		return fmt.Errorf("UserRepo - CreateUser - r.db.Exec: %w", err)
	}

	return nil
}

// GetUserByID retrieves a user by ID
func (r *UserRepo) GetUserByID(ctx context.Context, id string) (*entity.User, error) {
	query := `SELECT id, username, email, password, role, created_at, updated_at
	          FROM users WHERE id = ?`

	var user entity.User
	row := r.db.QueryRow(ctx, query, id)
	err := scanRow(row,
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, fmt.Errorf("usecase: user not found")
		}
		return nil, fmt.Errorf("UserRepo - GetUserByID - r.db.QueryRow: %w", err)
	}

	return &user, nil
}

// GetUserByUsername retrieves a user by username
func (r *UserRepo) GetUserByUsername(ctx context.Context, username string) (*entity.User, error) {
	query := `SELECT id, username, email, password, role, created_at, updated_at
	          FROM users WHERE username = ?`

	var user entity.User
	row := r.db.QueryRow(ctx, query, username)
	err := scanRow(row,
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, fmt.Errorf("usecase: user not found")
		}
		return nil, fmt.Errorf("UserRepo - GetUserByUsername - r.db.QueryRow: %w", err)
	}

	return &user, nil
}

// GetUserByEmail retrieves a user by email
func (r *UserRepo) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	query := `SELECT id, username, email, password, role, created_at, updated_at
	          FROM users WHERE email = ?`

	var user entity.User
	row := r.db.QueryRow(ctx, query, email)
	err := scanRow(row,
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, fmt.Errorf("usecase: user not found")
		}
		return nil, fmt.Errorf("UserRepo - GetUserByEmail - r.db.QueryRow: %w", err)
	}

	return &user, nil
}

// UpdateUser updates a user's information
func (r *UserRepo) UpdateUser(ctx context.Context, user entity.User) error {
	query := `UPDATE users SET username=?, email=?, role=?, updated_at=? WHERE id=?`

	err := r.db.Exec(ctx, query,
		user.Username,
		user.Email,
		user.Role,
		time.Now(),
		user.ID,
	)
	if err != nil {
		return fmt.Errorf("UserRepo - UpdateUser - r.db.Exec: %w", err)
	}

	return nil
}

// DeleteUser deletes a user
func (r *UserRepo) DeleteUser(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = ?`

	err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("UserRepo - DeleteUser - r.db.Exec: %w", err)
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
	query := `UPDATE users SET password=?, updated_at=? WHERE id=?`

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("UserRepo - UpdatePassword - bcrypt.GenerateFromPassword: %w", err)
	}

	err = r.db.Exec(ctx, query,
		string(hashedPassword),
		time.Now(),
		userID,
	)
	if err != nil {
		return fmt.Errorf("UserRepo - UpdatePassword - r.db.Exec: %w", err)
	}

	return nil
}
