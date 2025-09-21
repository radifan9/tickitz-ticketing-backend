package repositories

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/radifan9/tickitz-ticketing-backend/internal/models"
	"github.com/radifan9/tickitz-ticketing-backend/internal/utils"
	"github.com/redis/go-redis/v9"
)

type OrderRepository struct {
	db    *pgxpool.Pool
	rdb   *redis.Client
	cache *utils.CacheManager
}

func NewOrderRepository(db *pgxpool.Pool, rdb *redis.Client) *OrderRepository {
	return &OrderRepository{
		db:    db,
		rdb:   rdb,
		cache: utils.NewCacheManager(rdb),
	}
}

// Method used in Payment Page, when user clicked "Check Payment"
func (o *OrderRepository) AddNewTransactionsAndSeatCodes(ctx context.Context, t models.AddTransaction, userID string) (models.Transaction, error) {
	// Begin transaction
	tx, err := o.db.Begin(ctx)
	if err != nil {
		return models.Transaction{}, err
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				log.Println("failed to rollback transaction: ", rollbackErr)
			}
		}
	}()

	// Step 1: Insert seat codes and get their IDS
	var insertedSeatIDs []int
	if len(t.Seats) > 0 {
		insertedSeatIDs, err = o.insertSeatCodes(ctx, tx, t.Seats)
		if err != nil {
			return models.Transaction{}, err
		}
	}

	// Step 2: Insert new transaction
	newT, err := o.insertTransaction(ctx, tx, t, userID)
	if err != nil {
		return models.Transaction{}, err
	}

	// Step 3: Link seats to transaction
	if len(insertedSeatIDs) > 0 {
		err = o.linkSeatsToTransaction(ctx, tx, newT.ID, insertedSeatIDs)
		if err != nil {
			return models.Transaction{}, err
		}
	}

	// Commit the transaction
	if err = tx.Commit(ctx); err != nil {
		return models.Transaction{}, err
	}

	// Populate the rest of data from input (not from returning)
	newT.PaymentID = t.PaymentID
	newT.TotalPayment = t.TotalPayment
	newT.FullName = t.FullName
	newT.Email = t.Email
	newT.PhoneNumber = t.PhoneNumber
	newT.ScheduleID = t.ScheduleID
	newT.Seats = t.Seats

	return newT, nil

}

// Helper method to insert seat codes
func (o *OrderRepository) insertSeatCodes(ctx context.Context, tx pgx.Tx, seats []string) ([]int, error) {
	if len(seats) == 0 {
		return []int{}, nil
	}

	// Build query for adding seats
	placeholders := make([]string, len(seats))
	args := make([]interface{}, len(seats))
	for i, seat := range seats {
		placeholders[i] = fmt.Sprintf("($%d)", i+1)
		args[i] = seat
	}

	insertSeatsSQL := "INSERT INTO seat_codes (seat_code) VALUES " +
		strings.Join(placeholders, ",") + " RETURNING id"

	rows, err := tx.Query(ctx, insertSeatsSQL, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var insertedSeatIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		insertedSeatIDs = append(insertedSeatIDs, id)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return insertedSeatIDs, nil
}

// Helper method to insert transaction
func (o *OrderRepository) insertTransaction(ctx context.Context, tx pgx.Tx, t models.AddTransaction, userID string) (models.Transaction, error) {
	query := `
		INSERT INTO transactions (
			user_id,
			payment_id,
			total_payment,
			full_name,
			email,
			phone_number,
			schedule_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7) 
		RETURNING id::text, user_id::text, schedule_id`

	var newT models.Transaction
	err := tx.QueryRow(ctx, query,
		userID,
		t.PaymentID,
		t.TotalPayment,
		t.FullName,
		t.Email,
		t.PhoneNumber,
		t.ScheduleID,
	).Scan(
		&newT.ID,
		&newT.UserID,
		&newT.ScheduleID)

	if err != nil {
		return models.Transaction{}, err
	}

	return newT, nil
}

// Helper method to link seats to transaction
func (o *OrderRepository) linkSeatsToTransaction(ctx context.Context, tx pgx.Tx, transactionID string, seatIDs []int) error {
	if len(seatIDs) == 0 {
		return nil
	}

	// Build query for adding seat_code IDs to transactions_seats table
	placeholders := make([]string, len(seatIDs))
	args := make([]interface{}, len(seatIDs)*2) // *2 because we have transactionID and seatID for each row

	for i, seatID := range seatIDs {
		placeholders[i] = fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2)
		args[i*2] = transactionID
		args[i*2+1] = seatID
	}

	insertTSQuery := "INSERT INTO transactions_seats (transactions_id, seats_id) VALUES " +
		strings.Join(placeholders, ",")

	_, err := tx.Exec(ctx, insertTSQuery, args...)
	return err
}

// Patch transaction into paid by adding paid_at
func (o *OrderRepository) PayTransaction(ctx context.Context, transactionID string) (string, error) {
	query := `
		UPDATE transactions
		SET 
			paid_at = CURRENT_TIMESTAMP,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND paid_at IS NULL
		returning id
	`
	var id string
	if err := o.db.QueryRow(ctx, query, transactionID).Scan(&id); err != nil {
		return "", err
	}

	keysToInvalidate := []string{
		"tickitz:popular",
	}
	for _, k := range keysToInvalidate {
		if delErr := o.rdb.Del(ctx, k).Err(); delErr != nil {
			log.Printf("failed to invalidate cache for key %s: %v", k, delErr)
		}
	}

	return id, nil
}

// Transaction History
func (o *OrderRepository) ListTransaction(ctx context.Context, userID string) ([]models.TransactionHistory, error) {
	query := `
		select 
			t.id,
			c.name as cinema, 
			c.img as cinema_img, 
			s.show_date, 
			m.title,
			ar.age_rating, 
			st.start_at, 
			array_agg(seat_code) as seats ,
			t.total_payment, 
			t.phone_number, 
			t.paid_at, 
			t.updated_at,
			t.scanned_at, 
			t.schedule_id
		from transactions t
			join schedules s on t.schedule_id = s.id
			join movies m on s.movie_id = m.id
			join cinemas c on s.cinema_id = c.id
			join age_ratings ar on m.age_rating_id = ar.id
			join show_times st on s.show_time_id = st.id
			join transactions_seats ts on t.id = ts.transactions_id
			join seat_codes sc on ts.seats_id = sc.id
		where t.user_id = $1
		group by t.id, c.name, c.img, s.show_date, m.title,
			ar.age_rating, st.start_at, 
			t.total_payment, t.phone_number, 
			t.paid_at, t.scanned_at, t.schedule_id
		order by t.updated_at desc
	`
	rows, err := o.db.Query(ctx, query, userID)
	if err != nil {
		return []models.TransactionHistory{}, err
	}

	var listTransaction []models.TransactionHistory
	for rows.Next() {
		var t models.TransactionHistory
		if err := rows.Scan(
			&t.ID,
			&t.Cinema,
			&t.CinemaImg,
			&t.ShowDate,
			&t.Title,
			&t.AgeRating,
			&t.StartAt,
			&t.Seats,
			&t.TotalPayment,
			&t.PhoneNumber,
			&t.PaidAt,
			&t.UpdatedAt,
			&t.ScannedAt,
			&t.ScheduleID,
		); err != nil {
			return []models.TransactionHistory{}, err
		}
		listTransaction = append(listTransaction, t)
	}
	return listTransaction, nil
}
