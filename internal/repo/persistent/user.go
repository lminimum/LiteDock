package persistent

import (
	"context"
	"time"

	"github.com/lminimum/LiteDock/internal/entity"
	"github.com/lminimum/LiteDock/pkg/database"
	"github.com/lminimum/LiteDock/pkg/errors"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserRepo struct {
	db database.DB
}

func NewUserRepo(db database.DB) *UserRepo {
	return &UserRepo{db: db}
}

func scanRow(row interface{}, dest ...interface{}) error {
	if row == nil {
		return errors.ErrScanRowNil
	}
	scanner, ok := row.(interface{ Scan(...interface{}) error })
	if !ok {
		return errors.ErrScanRowNoScanner
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

func (r *UserRepo) CreateUser(ctx context.Context, user entity.User) error {
	query := `
		INSERT INTO users(id, username, email, password, role, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.Wrap(err, "UserRepo.CreateUser.Bcrypt")
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
		return errors.Wrap(err, "UserRepo.CreateUser.Exec")
	}

	return nil
}

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
			return nil, errors.Wrap(errors.ErrUserNotFound, "UserRepo.GetUserByID")
		}
		return nil, errors.Wrap(err, "UserRepo.GetUserByID.QueryRow")
	}

	return &user, nil
}

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
			return nil, errors.Wrap(errors.ErrUserNotFound, "UserRepo.GetUserByUsername")
		}
		return nil, errors.Wrap(err, "UserRepo.GetUserByUsername.QueryRow")
	}

	return &user, nil
}

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
			return nil, errors.Wrap(errors.ErrUserNotFound, "UserRepo.GetUserByEmail")
		}
		return nil, errors.Wrap(err, "UserRepo.GetUserByEmail.QueryRow")
	}

	return &user, nil
}

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
		return errors.Wrap(err, "UserRepo.UpdateUser.Exec")
	}

	return nil
}

func (r *UserRepo) DeleteUser(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = ?`

	err := r.db.Exec(ctx, query, id)
	if err != nil {
		return errors.Wrap(err, "UserRepo.DeleteUser.Exec")
	}

	return nil
}

func (r *UserRepo) VerifyPassword(ctx context.Context, username, password string) (bool, error) {
	user, err := r.GetUserByUsername(ctx, username)
	if err != nil {
		return false, errors.Wrap(err, "UserRepo.VerifyPassword.GetUserByUsername")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return false, errors.Wrap(err, "UserRepo.VerifyPassword.Bcrypt")
	}

	return true, nil
}

func (r *UserRepo) CountUsers(ctx context.Context) (int64, error) {
	var count int64

	row := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM users")
	err := scanRow(row, &count)

	return count, err
}

func (r *UserRepo) UpdatePassword(ctx context.Context, userID, newPassword string) error {
	query := `UPDATE users SET password=?, updated_at=? WHERE id=?`

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.Wrap(err, "UserRepo.UpdatePassword.Bcrypt")
	}

	err = r.db.Exec(ctx, query,
		string(hashedPassword),
		time.Now(),
		userID,
	)
	if err != nil {
		return errors.Wrap(err, "UserRepo.UpdatePassword.Exec")
	}

	return nil
}
