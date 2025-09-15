package models

import (
	"mime/multipart"
	"time"
)

type Movie struct {
	ID              int       `db:"id" json:"id"`
	Title           string    `db:"title" json:"title"`
	Synopsis        string    `db:"synopsis" json:"synopsis,omitempty"`
	PosterImg       string    `db:"poster_img" json:"poster_img"`
	BackdropImg     string    `db:"backdrop_img" json:"backdrop_img,omitempty"`
	DurationMinutes int       `db:"duration_minutes" json:"duration_minutes,omitempty"`
	ReleaseDate     time.Time `db:"release_date" json:"release_date"`
	Genres          []string  `db:"genres" json:"genres"`
	Director        string    `db:"director" json:"director,omitempty"`
	Cast            []string  `db:"cast" json:"cast,omitempty"`
}

// type CreateMovie struct {
// 	ID              int       `db:"id" json:"id"`
// 	Title           string    `db:"title" json:"title"`
// 	Synopsis        string    `db:"synopsis" json:"synopsis,omitempty"`
// 	PosterImg       string    `db:"poster_img" json:"poster_img"`
// 	BackdropImg     string    `db:"backdrop_img" json:"backdrop_img,omitempty"`
// 	DurationMinutes int       `db:"duration_minutes" json:"duration_minutes,omitempty"`
// 	ReleaseDate     time.Time `db:"release_date" json:"release_date"`
// 	Genres          []string  `db:"genres" json:"genres"`
// 	Director        string    `db:"director" json:"director,omitempty"`
// 	Cast            []string  `db:"cast" json:"cast,omitempty"`
// 	AgeRating       int       `json:"age_rating_id"`

// 	// Create Schedule

// }

type CreateMovie struct {
	ID              int                   `db:"id" form:"id"`
	Title           string                `db:"title" form:"title" json:"title,omitempty"`
	Synopsis        string                `db:"synopsis" form:"synopsis,omitempty" json:"synopsis,omitempty"`
	PosterImg       *multipart.FileHeader `db:"poster_img" form:"poster_img" json:"poster_img,omitempty"`
	BackdropImg     *multipart.FileHeader `db:"backdrop_img" form:"backdrop_img,omitempty" json:"backdrop_img,omitempty"`
	DurationMinutes int                   `db:"duration_minutes" form:"duration_minutes,omitempty" json:"duration_minutes,omitempty"`
	ReleaseDate     time.Time             `db:"release_date" form:"release_date" json:"release_date,omitempty"`
	Genres          string                `db:"genres" form:"genres" json:"genres,omitempty"`
	Director        string                `db:"director" form:"director,omitempty" json:"director,omitempty"`
	Cast            string                `db:"cast" form:"cast,omitempty" json:"cast,omitempty"`
	AgeRating       int                   `form:"age_rating_id" json:"age_rating_id,omitempty"`

	// Create Schedule

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
