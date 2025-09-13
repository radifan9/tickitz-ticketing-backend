package repositories

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/radifan9/tickitz-ticketing-backend/internal/models"
)

type ScheduleRepository struct {
	db *pgxpool.Pool
}

func NewScheduleRepository(db *pgxpool.Pool) *ScheduleRepository {
	return &ScheduleRepository{db: db}
}

func (s *ScheduleRepository) FilterSchedule(ctx context.Context, queryParam models.ScheduleFilter) ([]models.Schedule, error) {
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

	rows, err := s.db.Query(ctx, query, queryParam.MovieID, queryParam.CityID, queryParam.ShowTimeID, queryParam.Date)
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
			// return nil, err
			return []models.Schedule{}, err
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
func (s *ScheduleRepository) GetSoldSeatsByScheduleID(ctx context.Context, scheduleID string) ([]string, error) {
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
	if err := s.db.QueryRow(ctx, query, scheduleID).Scan(&seatCodes); err != nil {
		return nil, err
	}
	return seatCodes, nil
}
