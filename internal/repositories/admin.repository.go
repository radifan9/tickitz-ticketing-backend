package repositories

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AdminRepository struct {
	db *pgxpool.Pool
}

func (a *AdminRepository) MovieSales(ctx context.Context, movieID int) {

}
