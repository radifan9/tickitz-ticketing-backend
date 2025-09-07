package repositories

import (
	"context"
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
