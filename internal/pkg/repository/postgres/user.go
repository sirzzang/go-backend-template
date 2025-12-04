package postgres

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/your-org/go-backend-template/internal/pkg/entity"
)

// InsertUser creates a new user and returns the created user ID.
func (r *Repository) InsertUser(user *entity.User) (int, error) {
	ctx, cancel := r.GetContext()
	defer cancel()

	query := `
		INSERT INTO users (email, username, password, name, role, is_active)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	var id int
	err := r.db.QueryRowContext(ctx, query,
		user.Email,
		user.Username,
		user.Password,
		user.Name,
		user.Role,
		user.IsActive,
	).Scan(&id)

	if err != nil {
		// Check for unique constraint violation
		if strings.Contains(err.Error(), "unique constraint") ||
			strings.Contains(err.Error(), "duplicate key") {
			return 0, ErrDuplicateEmail
		}
		return 0, err
	}

	return id, nil
}

// GetUserById retrieves a user by ID.
func (r *Repository) GetUserById(id int) (*entity.User, error) {
	ctx, cancel := r.GetContext()
	defer cancel()

	query := `
		SELECT id, email, username, password, name, role, is_active, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	user := &entity.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.Id,
		&user.Email,
		&user.Username,
		&user.Password,
		&user.Name,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNoUser
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserByEmail retrieves a user by email.
func (r *Repository) GetUserByEmail(email string) (*entity.User, error) {
	ctx, cancel := r.GetContext()
	defer cancel()

	query := `
		SELECT id, email, username, password, name, role, is_active, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	user := &entity.User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.Id,
		&user.Email,
		&user.Username,
		&user.Password,
		&user.Name,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNoUser
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetUsers retrieves users with pagination.
func (r *Repository) GetUsers(offset, limit int, onlyActive bool) ([]*entity.User, error) {
	ctx, cancel := r.GetContext()
	defer cancel()

	query := `
		SELECT id, email, username, password, name, role, is_active, created_at, updated_at
		FROM users
		WHERE ($1 = false OR is_active = true)
		ORDER BY created_at DESC
		OFFSET $2
	`

	args := []any{onlyActive, offset}

	if limit > 0 {
		query += " LIMIT $3"
		args = append(args, limit)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]*entity.User, 0)
	for rows.Next() {
		user := &entity.User{}
		if err := rows.Scan(
			&user.Id,
			&user.Email,
			&user.Username,
			&user.Password,
			&user.Name,
			&user.Role,
			&user.IsActive,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, rows.Err()
}

// GetUserCount returns the total number of users.
func (r *Repository) GetUserCount(onlyActive bool) (int, error) {
	ctx, cancel := r.GetContext()
	defer cancel()

	query := `SELECT COUNT(*) FROM users WHERE ($1 = false OR is_active = true)`

	var count int
	if err := r.db.QueryRowContext(ctx, query, onlyActive).Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

// UpdateUser updates an existing user.
func (r *Repository) UpdateUser(user *entity.User) error {
	ctx, cancel := r.GetContext()
	defer cancel()

	query := `
		UPDATE users
		SET email = $1, username = $2, name = $3, role = $4, is_active = $5, updated_at = CURRENT_TIMESTAMP
		WHERE id = $6
	`

	result, err := r.db.ExecContext(ctx, query,
		user.Email,
		user.Username,
		user.Name,
		user.Role,
		user.IsActive,
		user.Id,
	)
	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") ||
			strings.Contains(err.Error(), "duplicate key") {
			return ErrDuplicateEmail
		}
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNoUser
	}

	return nil
}

// UpdateUserPassword updates a user's password.
func (r *Repository) UpdateUserPassword(id int, hashedPassword string) error {
	ctx, cancel := r.GetContext()
	defer cancel()

	query := `
		UPDATE users
		SET password = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`

	result, err := r.db.ExecContext(ctx, query, hashedPassword, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNoUser
	}

	return nil
}

// DeleteUserById deletes a user by ID.
func (r *Repository) DeleteUserById(id int) error {
	ctx, cancel := r.GetContext()
	defer cancel()

	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNoUser
	}

	return nil
}

// ExistsUserByEmail checks if a user with the given email exists.
func (r *Repository) ExistsUserByEmail(email string) (bool, error) {
	ctx, cancel := r.GetContext()
	defer cancel()

	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`

	var exists bool
	if err := r.db.QueryRowContext(ctx, query, email).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}

