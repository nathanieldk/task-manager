package repository

import (
	"context"
	"database/sql"

	"github.com/nathanieldk/task-manager/internal/entity"
)

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
	query := `
		INSERT INTO users (id, email, username, password_hash, team_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.db.ExecContext(ctx, query,
		user.ID, user.Email, user.Username, user.PasswordHash,
		user.TeamID, user.CreatedAt, user.UpdatedAt,
	)
	return err
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*entity.User, error) {
	query := `
		SELECT id, email, username, password_hash, team_id, created_at, updated_at
		FROM users WHERE id = $1`

	user := &entity.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.Username, &user.PasswordHash,
		&user.TeamID, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	query := `
		SELECT id, email, username, password_hash, team_id, created_at, updated_at
		FROM users WHERE email = $1`

	user := &entity.User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.Username, &user.PasswordHash,
		&user.TeamID, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *userRepository) FindByUsername(ctx context.Context, username string) (*entity.User, error) {
	query := `
		SELECT id, email, username, password_hash, team_id, created_at, updated_at
		FROM users WHERE username = $1`

	user := &entity.User{}
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID, &user.Email, &user.Username, &user.PasswordHash,
		&user.TeamID, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}
