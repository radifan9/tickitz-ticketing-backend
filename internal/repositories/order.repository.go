package repositories

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/radifan9/tickitz-ticketing-backend/internal/models"
)

type OrderRepository struct {
	db *pgxpool.Pool
}

func NewOrderRepository(db *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{db: db}
}

// --- Method used in Payment Page, when user clicked "Check Payment"
func (o *OrderRepository) AddNewTransactionsAndSeatCodes(ctx context.Context, t models.Transaction) (models.Transaction, error) {
	// --- --- Build Query for adding seats
	placeholders := make([]string, len(t.Seats))
	args := make([]interface{}, len(t.Seats))
	for i, s := range t.Seats {
		placeholders[i] = fmt.Sprintf("($%d)", i+1)
		args[i] = s
	}
	insertSeatsSQL := "insert into seat_codes (seat_code) values " + strings.Join(placeholders, ",") + " returning id"
	rows, err := o.db.Query(ctx, insertSeatsSQL, args...)
	if err != nil {
		return models.Transaction{}, err
	}

	var insertedSeatIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return models.Transaction{}, err
		}
		insertedSeatIDs = append(insertedSeatIDs, id)
	}

	log.Println(insertedSeatIDs)

	// --- --- Build Query for add new transaction (Assuming user paid directly)
	query := `
		insert into
	  transactions (
	    user_id,
	    payment_id,
	    total_payment,
	    full_name,
	    email,
	    phone_number,
	    paid_at,
	    schedule_id
	  )
	values
	  ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP, $7) 
		returning id:: text, user_id:: text, schedule_id
		`

	var newT models.Transaction
	err2 := o.db.QueryRow(ctx, query,
		t.UserID,
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
	if err2 != nil {
		return models.Transaction{}, err2
	}

	// Populate the rest of data from input (not from returning)
	newT.PaymentID = t.PaymentID
	newT.TotalPayment = t.TotalPayment
	newT.FullName = t.FullName
	newT.Email = t.Email
	newT.PhoneNumber = t.PhoneNumber
	newT.ScheduleID = t.ScheduleID
	newT.Seats = t.Seats

	// --- --- Build Query for adding seat_code IDs to transactions_seats table (assoc table)
	placeholders3 := make([]string, len(insertedSeatIDs))
	args3 := make([]interface{}, len(insertedSeatIDs))
	for i, sID := range insertedSeatIDs {
		placeholders3[i] = fmt.Sprintf("('%s', $%d)", newT.ID, i+1)
		args3[i] = sID
	}
	insertTSQuery := "insert into transactions_seats (transactions_id, seats_id) values " + strings.Join(placeholders3, ",")

	_, err3 := o.db.Query(ctx, insertTSQuery, args3...)
	if err3 != nil {
		return models.Transaction{}, err3
	}

	return newT, nil
}

// --- Transaction History
func (o *OrderRepository) ListTransaction(ctx context.Context, userID string) ([]models.TransactionHistory, error) {
	query := `
		select 
			t.id,
			c.name as cinema, c.img as cinema_img, s.show_date, m.title,
			ar.age_rating, st.start_at, array_agg(seat_code) as seats ,
			t.total_payment, t.phone_number, 
			t.paid_at, t.scanned_at, t.schedule_id
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
			&t.ScannedAt,
			&t.ScheduleID,
		); err != nil {
			return []models.TransactionHistory{}, err
		}
		listTransaction = append(listTransaction, t)
	}
	return listTransaction, nil
}
