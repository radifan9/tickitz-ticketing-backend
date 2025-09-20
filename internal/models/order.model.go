package models

import "time"

type Schedule struct {
	ID         int    `db:"id" json:"schedule_id"`
	MovieID    int    `db:"movie_id" json:"movie_id"`
	Title      string `db:"title" json:"title"`
	ShowDate   string `db:"show_date" json:"show_date"`
	ShowTimeID int    `db:"show_time_id" json:"show_time_id"`
	StartAt    string `db:"start_at" json:"start_at"`
	CityID     int    `db:"city_id" json:"city_id"`
	CityName   string `db:"city_name" json:"city_name"`
	CinemaID   int    `db:"cinema_id" json:"cinema_id"`
	CinemaName string `db:"cinema_name" json:"cinema_name"`
	CinemaImg  string `db:"img" json:"cinema_img"`
}

type ScheduleFilter struct {
	MovieID    string `form:"movie_id"`
	CityID     string `form:"city_id"`
	ShowTimeID string `form:"show_time_id"`
	Date       string `form:"show_date"`
}

type AddTransaction struct {
	ID           string     `json:"id,omitempty"`
	UserID       string     `json:"user_id,omitempty"`
	PaymentID    int        `json:"payment_id,omitempty"`
	TotalPayment int        `json:"total_payment"`
	FullName     string     `json:"full_name"`
	Email        string     `json:"email"`
	PhoneNumber  string     `json:"phone_number"`
	CreatedAt    *time.Time `json:"created_at,omitempty"`
	ScheduleID   int        `json:"schedule_id"`
	Seats        []string   `json:"seats"`
}

type Transaction struct {
	ID           string     `json:"id,omitempty"`
	UserID       string     `json:"user_id,omitempty"`
	PaymentID    int        `json:"payment_id,omitempty"`
	TotalPayment int        `json:"total_payment"`
	FullName     string     `json:"full_name"`
	Email        string     `json:"email"`
	PhoneNumber  string     `json:"phone_number"`
	PaidAt       *time.Time `json:"paid_at,omitempty"`
	ScannedAt    *time.Time `json:"scanned_at,omitempty"`
	CreatedAt    *time.Time `json:"created_at,omitempty"`
	UpdatedAt    *time.Time `json:"updated_at,omitempty"`
	ScheduleID   int        `json:"schedule_id"`
	Seats        []string   `json:"seats"`
}

type TransactionHistory struct {
	Transaction
	Cinema    string    `json:"cinema"`
	CinemaImg string    `json:"cinema_img"`
	ShowDate  time.Time `json:"show_date"`
	Title     string    `json:"title"`
	AgeRating string    `json:"age_rating"`
	StartAt   string    `json:"start_at"`
}

type SeatCodes struct {
	ID        int
	seat_code string
}
