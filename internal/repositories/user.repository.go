package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/radifan9/tickitz-ticketing-backend/internal/models"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (u *UserRepository) CreateUser(ctx context.Context, email, hashedPassword string) (models.User, error) {
	query := `INSERT INTO users (email, password) VALUES ($1, $2) RETURNING id, email`
	var user models.User

	if err := u.db.QueryRow(ctx, query, email, hashedPassword).Scan(&user.Id, &user.Email); err != nil {
		return models.User{}, fmt.Errorf("failed to register user: %w", err)
	}
	return user, nil
}

func (u *UserRepository) GetIDFromEmail(ctx context.Context, email string) (models.User, error) {
	query := `SELECT id FROM users WHERE email = $1`

	var user models.User

	if err := u.db.QueryRow(ctx, query, email).Scan(&user.Id); err != nil {
		return models.User{}, errors.New("failed to login")
	}
	return user, nil
}

func (u *UserRepository) GetPasswordFromID(ctx context.Context, id string) (models.User, error) {
	query := `SELECT role, password FROM users WHERE id = $1`

	var user models.User

	if err := u.db.QueryRow(ctx, query, id).Scan(&user.Role, &user.Password); err != nil {
		return models.User{}, errors.New("failed to login")
	}
	return user, nil
}
