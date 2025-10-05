package models

import (
	"mime/multipart"
	"time"
)

type Movie struct {
	ID              int        `db:"id" json:"id"`
	Title           string     `db:"title" json:"title"`
	Synopsis        string     `db:"synopsis" json:"synopsis,omitempty"`
	PosterImg       string     `db:"poster_img" json:"poster_img"`
	BackdropImg     string     `db:"backdrop_img" json:"backdrop_img,omitempty"`
	DurationMinutes *int       `db:"duration_minutes" json:"duration_minutes,omitempty"`
	ReleaseDate     *time.Time `db:"release_date" json:"release_date,omitempty"`
	AgeRatingID     int        `json:"age_rating_id,omitempty"`
	Genres          []string   `db:"genres" json:"genres"`
	Director        string     `db:"director" json:"director,omitempty"`
	Cast            []string   `db:"cast" json:"cast,omitempty"`
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at" json:"updated_at"`
}

type CreateMovie struct {
	ID              int                   `db:"id" form:"id"`
	PosterImg       *multipart.FileHeader `db:"poster_img" form:"poster_img" json:"poster_img,omitempty"`
	BackdropImg     *multipart.FileHeader `db:"backdrop_img" form:"backdrop_img,omitempty" json:"backdrop_img,omitempty"`
	Title           string                `db:"title" form:"title" json:"title,omitempty"`
	Genres          string                `db:"genres" form:"genres" json:"genres,omitempty"`
	AgeRating       int                   `form:"age_rating_id" json:"age_rating_id,omitempty"`
	ReleaseDate     string                `db:"release_date" form:"release_date" json:"release_date,omitempty"`
	DurationMinutes int                   `db:"duration_minutes" form:"duration_minutes,omitempty" json:"duration_minutes,omitempty"`
	Director        string                `db:"director" form:"director,omitempty" json:"director,omitempty"`
	Cast            string                `db:"cast" form:"cast,omitempty" json:"cast,omitempty"`
	Synopsis        string                `db:"synopsis" form:"synopsis,omitempty" json:"synopsis,omitempty"`
	CityID          string                `db:"city_id" form:"city_id" json:"city_id"`
	CinemaID        string                `db:"cinema_id" form:"cinema_id" json:"cinema_id"`
	ShowDate        string                `db:"show_date" form:"show_date" json:"show_date"`
	ShowTimeID      string                `db:"show_time_id" form:"show_time_id" json:"show_time_id"`
}

type MovieFilter struct {
	Keywords []string `db:"keywords"`
	Genres   []int    `db:"genres"`
	Offset   int      `db:"offset"`
	Limit    int      `db:"limit"`
}

type ArchiveMovieRespond struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Archived_at time.Time `json:"archived_at"`
}
