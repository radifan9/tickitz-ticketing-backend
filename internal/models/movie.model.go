package models

import "time"

type Movie struct {
	ID              int       `db:"id" json:"id"`
	Title           string    `db:"title" json:"title"`
	Synopsis        string    `db:"synopsis" json:"synopsis,omitempty"`
	PosterImg       string    `db:"poster_img" json:"poster_img"`
	BackdropImg     string    `db:"backdrop_img" json:"backdrop_img,omitempty"`
	DurationMinutes int       `db:"duration_minutes" json:"duration_minutes,omitempty"`
	ReleaseDate     time.Time `db:"release_date" json:"release_date"`
	Genres          []string  `db:"genres" json:"genres"`
}

type MovieFilter struct {
	Keywords []string `db:"keywords"`
	Genres   []int    `db:"genres"`
	Offset   int      `db:"offset"`
	Limit    int      `db:"limit"`
}
