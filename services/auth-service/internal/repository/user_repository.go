package repository

import (
	"database/sql"
	"fmt"
	"time"

	"healthcare-platform/services/auth-service/internal/model"
)

// UserRepository handles all database operations for users
// Only this file should talk directly to PostgreSQL
type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// ──────────────────────────────────────────────
// User CRUD operations
// ──────────────────────────────────────────────

func (r *UserRepository) Create(user *model.User) error {
	query := `
		INSERT INTO users 
			(id, email, password_hash, role, first_name, last_name, is_verified, is_active, created_at, updated_at)
		VALUES 
			($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.Exec(query,
		user.ID,
		user.Email,
		user.PasswordHash,
		string(user.Role),
		user.FirstName,
		user.LastName,
		user.IsVerified,
		user.IsActive,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("repository.Create: %w", err)
	}
	return nil
}

func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	query := `
		SELECT id, email, password_hash, role, first_name, last_name, 
		       is_verified, is_active, created_at, updated_at
		FROM users
		WHERE email = $1 AND is_active = TRUE
	`
	return r.scanUser(r.db.QueryRow(query, email))
}

func (r *UserRepository) FindByID(id string) (*model.User, error) {
	query := `
		SELECT id, email, password_hash, role, first_name, last_name, 
		       is_verified, is_active, created_at, updated_at
		FROM users
		WHERE id = $1 AND is_active = TRUE
	`
	return r.scanUser(r.db.QueryRow(query, id))
}

func (r *UserRepository) EmailExists(email string) (bool, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM users WHERE email = $1", email).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("repository.EmailExists: %w", err)
	}
	return count > 0, nil
}

func (r *UserRepository) SetVerified(userID string) error {
	_, err := r.db.Exec(
		"UPDATE users SET is_verified = TRUE, updated_at = NOW() WHERE id = $1",
		userID,
	)
	return err
}

func (r *UserRepository) SetActive(userID string, active bool) error {
	_, err := r.db.Exec(
		"UPDATE users SET is_active = $1, updated_at = NOW() WHERE id = $2",
		active, userID,
	)
	return err
}

// ──────────────────────────────────────────────
// Refresh token operations
// ──────────────────────────────────────────────

func (r *UserRepository) SaveRefreshToken(userID, tokenHash string, expiresAt time.Time) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`
	_, err := r.db.Exec(query, userID, tokenHash, expiresAt)
	if err != nil {
		return fmt.Errorf("repository.SaveRefreshToken: %w", err)
	}
	return nil
}

// FindRefreshToken returns userID if token is valid and not expired
func (r *UserRepository) FindRefreshToken(tokenHash string) (string, error) {
	var userID string
	err := r.db.QueryRow(
		"SELECT user_id FROM refresh_tokens WHERE token_hash = $1 AND expires_at > NOW()",
		tokenHash,
	).Scan(&userID)

	if err == sql.ErrNoRows {
		return "", nil // Not found — return empty, no error
	}
	if err != nil {
		return "", fmt.Errorf("repository.FindRefreshToken: %w", err)
	}
	return userID, nil
}

func (r *UserRepository) DeleteRefreshToken(tokenHash string) error {
	_, err := r.db.Exec("DELETE FROM refresh_tokens WHERE token_hash = $1", tokenHash)
	return err
}

func (r *UserRepository) DeleteAllUserRefreshTokens(userID string) error {
	_, err := r.db.Exec("DELETE FROM refresh_tokens WHERE user_id = $1", userID)
	return err
}

// ──────────────────────────────────────────────
// Private helpers
// ──────────────────────────────────────────────

func (r *UserRepository) scanUser(row *sql.Row) (*model.User, error) {
	user := &model.User{}
	var roleStr string

	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&roleStr,
		&user.FirstName,
		&user.LastName,
		&user.IsVerified,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil // Not found
	}
	if err != nil {
		return nil, fmt.Errorf("repository.scanUser: %w", err)
	}

	user.Role = model.Role(roleStr)
	return user, nil
}
