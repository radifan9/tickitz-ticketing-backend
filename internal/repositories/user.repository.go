package repositories

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/radifan9/tickitz-ticketing-backend/internal/models"
	"github.com/redis/go-redis/v9"
)

type UserRepository struct {
	db  *pgxpool.Pool
	rdb *redis.Client
}

func NewUserRepository(db *pgxpool.Pool, rdb *redis.Client) *UserRepository {
	return &UserRepository{
		db:  db,
		rdb: rdb,
	}
}

func (u *UserRepository) CreateUser(ctx context.Context, email, hashedPassword string) (models.User, error) {
	// Begin transaction
	tx, err := u.db.Begin(ctx)
	if err != nil {
		return models.User{}, err
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				log.Println("failed to rollback transaction: ", rollbackErr)
			}
		}
	}()

	// Step 1: Create user

	query := `INSERT INTO users (email, password) VALUES ($1, $2) RETURNING id, email`
	var user models.User

	if err := u.db.QueryRow(ctx, query, email, hashedPassword).Scan(&user.Id, &user.Email); err != nil {
		return models.User{}, fmt.Errorf("failed to register user: %w", err)
	}

	// Step 2: Create Profile
	var profileID string
	profileID, err = u.createProfile(ctx, tx, user.Id)
	if err != nil {
		return models.User{}, err
	}

	log.Println("profile, userID : ", profileID)

	// Commit transaction if everything succeeds
	if err = tx.Commit(ctx); err != nil {
		return models.User{}, err
	}

	return user, nil
}

func (u *UserRepository) createProfile(ctx context.Context, tx pgx.Tx, userID string) (string, error) {
	query := `
		insert into user_profiles (user_id) values ($1) returning user_id`

	var id string
	err := tx.QueryRow(ctx, query, userID).Scan(&id)
	if err != nil {
		return "", err
	}

	return id, nil
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

// GetProfileByUserID fetches a user's profile from user_profiles by user_id
func (u *UserRepository) GetProfile(ctx context.Context, userID string) (models.UserProfile, error) {
	query := `
			SELECT 
					user_id,
					COALESCE(first_name, ''),
					COALESCE(last_name, ''),
					COALESCE(img, ''),
					COALESCE(phone_number, ''),
					COALESCE(points, 0)
			FROM user_profiles
			WHERE user_id = $1
	`

	var p models.UserProfile
	if err := u.db.QueryRow(ctx, query, userID).Scan(
		&p.UserID,
		&p.FirstName,
		&p.LastName,
		&p.Img,
		&p.PhoneNumber,
		&p.Points,
	); err != nil {
		return models.UserProfile{}, fmt.Errorf("profile not found or error fetching profile: %w", err)
	}
	return p, nil
}

func (u *UserRepository) EditProfile(ctx context.Context, userID string, body models.EditUserProfile, imagePath string) (models.UserProfile, error) {
	sql := "UPDATE user_profiles SET "
	values := []any{}

	if body.FirstName != "" {
		sql += fmt.Sprintf("%s=$%d, ", "first_name", len(values)+1)
		values = append(values, body.FirstName)
	}
	if body.LastName != "" {
		sql += fmt.Sprintf("%s=$%d, ", "last_name", len(values)+1)
		values = append(values, body.LastName)
	}
	if body.PhoneNumber != "" {
		sql += fmt.Sprintf("%s=$%d, ", "phone_number", len(values)+1)
		values = append(values, body.PhoneNumber)
	}
	// if you decide to save image filename:
	if body.Img != nil {
		sql += fmt.Sprintf("%s=$%d, ", "img", len(values)+1)
		values = append(values, imagePath)
	}

	// sql += fmt.Sprintf("updated_at=CURRENT_TIMESTAMP WHERE user_id=$%d RETURNING user_id, first_name, last_name, img, phone_number, points, created_at, updated_at", len(values)+1)
	sql += fmt.Sprintf(`updated_at=CURRENT_TIMESTAMP 
    WHERE user_id=$%d 
    RETURNING 
        user_id, 
        COALESCE(first_name, ''), 
        COALESCE(last_name, ''), 
        COALESCE(img, ''), 
        COALESCE(phone_number, ''), 
        COALESCE(points, 0), 
        created_at, 
        updated_at`, len(values)+1)

	values = append(values, userID)

	var profile models.UserProfile
	if err := u.db.QueryRow(ctx, sql, values...).Scan(
		&profile.UserID,
		&profile.FirstName,
		&profile.LastName,
		&profile.Img,
		&profile.PhoneNumber,
		&profile.Points,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	); err != nil {
		return models.UserProfile{}, err
	}

	return profile, nil
}

func (u *UserRepository) UpdatePassword(ctx context.Context, userID, hashedPassword string) error {
	query := `UPDATE users SET password = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	_, err := u.db.Exec(ctx, query, hashedPassword, userID)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}
	return nil
}
