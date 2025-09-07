package repositories

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/radifan9/tickitz-ticketing-backend/internal/models"
)

// Struct that holds shared dependency
type MovieRepository struct {
	db *pgxpool.Pool
}

// Constructor function
// Purpose: creates a repository instance in a valid state (db injected), returning a pointer to use its methods.
func NewMovieRepository(db *pgxpool.Pool) *MovieRepository {
	return &MovieRepository{db: db}
}

func (m *MovieRepository) ListUpcomingMovie(ctx context.Context) ([]models.Movie, error) {
	// Query for getting upcoming (not release yet) movies
	query := `
	SELECT
		m.id,
		m.title,
		m.poster_img,
		m.release_date,
		ARRAY_AGG(
			g.name
			ORDER BY
				g.name
		) AS genres
	FROM
		movies m
		JOIN movie_genres mg ON m.id = mg.movie_id
		JOIN genres g ON mg.genre_id = g.id
	WHERE
		m.release_date > CURRENT_DATE
	GROUP BY
		m.id,
		m.title,
		m.poster_img,
		m.release_date
	ORDER BY
		m.release_date ASC`
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

// Filter by keywords, genres, and pagination
func (m *MovieRepository) ListMovieFiltered(
	ctx context.Context,
	filter models.MovieFilter,
) ([]models.Movie, error) {

	// Base Query, this works even without any search query
	baseQuery := `
		SELECT
			m.id,
			m.title,
			m.synopsis,
			m.poster_img,
			m.backdrop_img,
			m.duration_minutes,
			m.release_date,
			ARRAY_AGG(DISTINCT g.name ORDER BY g.name) AS genres
		FROM movies m
		JOIN movie_genres mg ON m.id = mg.movie_id
		JOIN genres g ON mg.genre_id = g.id
	`

	conds := []string{}     // accumulates SQL condition snippets
	args := []interface{}{} // accumulates parameter values
	argPos := 1             // track the next $n (start at 1)

	// Use reflection to inspect struct fields
	v := reflect.ValueOf(filter)
	t := reflect.TypeOf(filter)

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i).Interface()

		switch field.Tag.Get("db") {
		case "keywords":
			if kws, ok := value.([]string); ok && len(kws) > 0 {
				conds = append(conds, fmt.Sprintf(`
					EXISTS (
						SELECT 1 
						FROM unnest($%d::text[]) kw
						WHERE m.title ILIKE '%%' || kw || '%%'
					)
				`, argPos))
				args = append(args, kws)
				argPos++
			}

		case "genres":
			if gs, ok := value.([]int); ok && len(gs) > 0 {
				conds = append(conds, fmt.Sprintf("g.id = ANY($%d::int[])", argPos))
				args = append(args, gs)
				argPos++
			}

			// offset & limit are handled later
		}
	}

	if len(conds) > 0 {
		baseQuery += " WHERE " + strings.Join(conds, " AND ")
	}

	// Final GROUP BY, ORDER, OFFSET/LIMIT
	baseQuery += fmt.Sprintf(`
		GROUP BY m.id, m.title, m.synopsis, m.poster_img, m.backdrop_img, m.duration_minutes, m.release_date
		ORDER BY m.release_date ASC
		OFFSET $%d LIMIT $%d
	`, argPos, argPos+1)

	args = append(args, filter.Offset, filter.Limit)

	// Run query
	rows, err := m.db.Query(ctx, baseQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Scan results
	var movies []models.Movie
	for rows.Next() {
		var movie models.Movie
		if err := rows.Scan(
			&movie.ID,
			&movie.Title,
			&movie.Synopsis,
			&movie.PosterImg,
			&movie.BackdropImg,
			&movie.DurationMinutes,
			&movie.ReleaseDate,
			&movie.Genres,
		); err != nil {
			return nil, err
		}
		movies = append(movies, movie)
	}

	return movies, nil
}
