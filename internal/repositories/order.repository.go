package repositories

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/radifan9/tickitz-ticketing-backend/internal/models"
)

type OrderRepository struct {
	db *pgxpool.Pool
}

func NewOrderRepository(db *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{db: db}
}

func (o *OrderRepository) FilterSchedule(ctx context.Context, movieID, cityID, showTimeID string) ([]models.Schedule, error) {
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
			WHERE s.show_date = CURRENT_DATE + INTERVAL '1 day'
				AND ($1::int IS NULL OR s.movie_id = $1)
				AND ($2::int IS NULL OR s.city_id = $2)
				AND ($3::int IS NULL OR s.show_time_id = $3)
			ORDER BY s.show_time_id, s.cinema_id;
	`

	rows, err := o.db.Query(ctx, query, movieID, cityID, showTimeID)
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

func (o *OrderRepository) AddNewTransactions(ctx context.Context, t models.Transaction) (models.Transaction, error) {
	// Assume user book and ticket and paid directly
	query := `
	insert into transactions
	(user_id, payment_id, paid_at, total_payment, is_paid, full_name, email, phone_number, schedule_id) values (
		$1, $2, CURRENT_TIMESTAMP, $3, $4, $5, $6, $7, $8
	) returning id::text, user_id::text
	`

	var newTransaction models.Transaction
	err := o.db.QueryRow(ctx, query,
		t.UserID,
		t.PaymentID,
		t.TotalPayment,
		t.IsPaid,
		t.FullName,
		t.Email,
		t.PhoneNumber,
		t.ScheduleID,
	).Scan(&newTransaction.ID, &newTransaction.UserID)
	if err != nil {
		return models.Transaction{}, err
	}

	// populate the rest from input
	newTransaction.PaymentID = t.PaymentID
	newTransaction.TotalPayment = t.TotalPayment
	newTransaction.FullName = t.FullName
	newTransaction.Email = t.Email
	newTransaction.PhoneNumber = t.PhoneNumber
	newTransaction.ScheduleID = t.ScheduleID
	newTransaction.IsPaid = t.IsPaid

	// If there's no error
	return newTransaction, nil
}
