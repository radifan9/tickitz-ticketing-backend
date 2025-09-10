package models

import "time"

type Schedule struct {
	ID         int    `db:"id" json:"id"`
	MovieID    int    `db:"movie_id" json:"movie_id"`
	Title      string `db:"title" json:"title"`
	CityID     int    `db:"city_id" json:"city_id"`
	CityName   string `db:"city_name" json:"city_name"`
	ShowTimeID int    `db:"show_time_id" json:"show_time_id"`
	StartAt    string `db:"start_at" json:"start_at"`
	CinemaID   int    `db:"cinema_id" json:"cinema_id"`
	CinemaName string `db:"cinema_name" json:"cinema_name"`
	CinemaImg  string `db:"img" json:"cinema_img"`
	ShowDate   string `db:"show_date" json:"show_date"`
}

type ScheduleFilter struct {
	MovieID    string `form:"movie_id"`
	CityID     string `form:"city_id"`
	ShowTimeID string `form:"show_time_id"`
	Date       string `form:"show_date"`
}

type Transaction struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	PaymentID    int       `json:"payment_id"`
	TotalPayment int       `json:"total_payment"`
	FullName     string    `json:"full_name"`
	Email        string    `json:"email"`
	PhoneNumber  string    `json:"phone_number"`
	PaidAt       time.Time `json:"paid_at"`
	ScannedAt    time.Time `json:"scanned_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	ScheduleID   int       `json:"schedule_id"`
	Seats        []string  `json:"seats"`
}

type SeatCodes struct {
	ID        int
	seat_code string
}
