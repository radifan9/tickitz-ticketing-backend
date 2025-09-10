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

func (o *OrderRepository) FilterSchedule(ctx context.Context, queryParam models.ScheduleFilter) ([]models.Schedule, error) {
	query := `
			SELECT 
				s.id,
				s.movie_id,
				m.title,
				s.city_id,
				ci.name AS city_name,
				s.show_time_id,
				st.start_at::text AS start_at,
				s.cinema_id,
				c.name AS cinema_name,
				c.img,
				s.show_date::text AS show_date
			FROM schedules s
			JOIN movies m ON s.movie_id = m.id
			JOIN cities ci ON s.city_id = ci.id
			JOIN show_times st ON s.show_time_id = st.id
			JOIN cinemas c ON s.cinema_id = c.id
			WHERE s.show_date = COALESCE(NULLIF($4, '')::date, CURRENT_DATE + 1)
				AND (NULLIF($1, '')::int IS NULL OR s.movie_id = NULLIF($1, '')::int)
				AND (NULLIF($2, '')::int IS NULL OR s.city_id = NULLIF($2, '')::int)
				AND (NULLIF($3, '')::int IS NULL OR s.show_time_id = NULLIF($3, '')::int)
			ORDER BY s.show_time_id, s.cinema_id;
	`

	rows, err := o.db.Query(ctx, query, queryParam.MovieID, queryParam.CityID, queryParam.ShowTimeID, queryParam.Date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []models.Schedule
	for rows.Next() {
		var s models.Schedule
		if err := rows.Scan(
			&s.ID, &s.MovieID, &s.Title,
			&s.CityID, &s.CityName,
			&s.ShowTimeID, &s.StartAt,
			&s.CinemaID, &s.CinemaName, &s.CinemaImg,
			&s.ShowDate,
		); err != nil {
			return nil, err
		}
		schedules = append(schedules, s)
	}
	return schedules, nil
}

// --- Method used in Choosing Seats
// In here user already choose schedule
// Logika
// Cari transactions dengan schedule ID yang dimasukkan, yang sudah dibayar
// --> Cari transactions_seats yang melekat pada transaction tersebut
// --> Cari seat_codes yang melekat pada transaksi tersebut
func (o *OrderRepository) GetSoldSeatsByScheduleID(ctx context.Context, scheduleID string) ([]string, error) {
	query := `
	WITH get_paid_transaction_id_by_schedule AS (
			SELECT id AS transaction_id
			FROM transactions
			WHERE schedule_id = $1
				AND paid_at IS NOT NULL
	),
	get_seats_id AS (
			SELECT seats_id AS seat_code_id
			FROM transactions_seats
			WHERE transactions_id IN (
					SELECT transaction_id FROM get_paid_transaction_id_by_schedule
			)
	)
	SELECT COALESCE(
			ARRAY_AGG(DISTINCT sc.seat_code ORDER BY sc.seat_code),
			'{}'
	) AS seat_codes
	FROM seat_codes sc
	WHERE sc.id IN (SELECT seat_code_id FROM get_seats_id);
	`

	var seatCodes []string
	if err := o.db.QueryRow(ctx, query, scheduleID).Scan(&seatCodes); err != nil {
		return nil, err
	}
	return seatCodes, nil
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
