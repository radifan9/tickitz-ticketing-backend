package repositories

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/radifan9/tickitz-ticketing-backend/internal/models"
)

type MovieRepository struct {
	db *pgxpool.Pool
}

func NewMovieRepository(db *pgxpool.Pool) *MovieRepository {
	return &MovieRepository{db: db}
}

func (m *MovieRepository) GetUpcomingMovie(ctx context.Context) ([]models.Movie, error) {
	query := `SELECT 
	m.id, 
	m.title, 
	m.poster_img, 
	m.release_date, 
	ARRAY_AGG(g.name ORDER BY g.name) AS genres 
	FROM movies m JOIN movie_genres mg ON m.id = mg.movie_id JOIN genres g ON mg.genre_id = g.id WHERE m.release_date > CURRENT_DATE GROUP BY m.id, m.title, m.poster_img, m.release_date ORDER BY m.release_date ASC`
	rows, err := m.db.Query(ctx, query)
	if err != nil {
		log.Println("internal server error : ", err.Error())
		return []models.Movie{}, err
	}
	defer rows.Close()

	var movies []models.Movie

	// Membaca rows/record
	for rows.Next() {
		var movie models.Movie
		if err := rows.Scan(&movie.ID, &movie.Title, &movie.PosterImg, &movie.ReleaseDate, &movie.Genres); err != nil {
			log.Println("scan error, ", err.Error())
			return []models.Movie{}, err
		}
		movies = append(movies, movie)
	}

	return movies, nil
}
