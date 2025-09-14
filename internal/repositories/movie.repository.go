package repositories

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/radifan9/tickitz-ticketing-backend/internal/models"
	"github.com/radifan9/tickitz-ticketing-backend/internal/utils"
	"github.com/redis/go-redis/v9"
)

// Struct that holds shared dependency
type MovieRepository struct {
	db    *pgxpool.Pool
	rdb   *redis.Client
	cache *utils.CacheManager
}

// Constructor function
// Purpose: creates a repository instance in a valid state (db injected), returning a pointer to use its methods.
func NewMovieRepository(db *pgxpool.Pool, rdb *redis.Client) *MovieRepository {
	return &MovieRepository{
		db:    db,
		rdb:   rdb,
		cache: utils.NewCacheManager(rdb),
	}
}

// --- List Upcoming Movies (not yet released current_date < release_dateI)
func (m *MovieRepository) ListUpcomingMovies(ctx context.Context) ([]models.Movie, error) {
	var movies []models.Movie
	redisKey := "tickitz:upcoming"

	// Use the cache utils with cache-aside pattern
	err := m.cache.CacheOrFetch(
		ctx,
		redisKey,
		24*time.Hour,
		&movies,
		func() (interface{}, error) {
			// This function will only be called on cache-miss
			return m.fetchUpcomingMoviesFromDB(ctx)
		},
	)

	return movies, err
}

// --- Fetching the Upcoming Movies
func (m *MovieRepository) fetchUpcomingMoviesFromDB(ctx context.Context) ([]models.Movie, error) {
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
			m.release_date ASC
	`

	rows, err := m.db.Query(ctx, query)
	if err != nil {
		return []models.Movie{}, err
	}
	defer rows.Close()

	var movies []models.Movie

	// Read rows/records
	for rows.Next() {
		var movie models.Movie
		if err := rows.Scan(&movie.ID, &movie.Title, &movie.PosterImg, &movie.ReleaseDate, &movie.Genres); err != nil {
			return []models.Movie{}, err
		}
		movies = append(movies, movie)
	}

	return movies, nil
}

// List Popular Movies
func (m *MovieRepository) ListPopularMovies(ctx context.Context) ([]models.Movie, error) {
	var movies []models.Movie
	rediskey := "tickitz:popular"

	err := m.cache.CacheOrFetch(
		ctx,
		rediskey,
		24*time.Hour,
		&movies,
		func() (interface{}, error) {
			return m.fetchPopularMovies(ctx)
		},
	)

	return movies, err
}

// Fetching the Popular Movie
func (m *MovieRepository) fetchPopularMovies(ctx context.Context) ([]models.Movie, error) {
	query := `
	with get_popular_movie_id as (
		select
			s.movie_id
		from
			transactions t
			join schedules s on t.schedule_id = s.id
		where
			-- Make sure the movie is still on the show_date schedule
			s.show_date >= current_date
		group by
			movie_id
		order by
			count(t.id) desc
	)
		
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
		m.id in (
			select
				movie_id
			from
				get_popular_movie_id
		)
	GROUP BY
		m.id,
		m.title,
		m.poster_img,
		m.release_date
	ORDER BY
		m.release_date ASC
	`

	rows, err := m.db.Query(ctx, query)
	if err != nil {
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

	// Check if its the first page of all movies
	if filter.Limit == 20 && filter.Offset == 0 {

		// If it's first page then try to get the cache
		var movies []models.Movie
		rediskey := "tickitz:movies-all-first-page"

		m.cache.CacheOrFetch(
			ctx,
			rediskey,
			24*time.Hour,
			&movies,
			func() (interface{}, error) {
				return m.fetchMovieFiltered(ctx, filter)
			},
		)

		return movies, nil
	} else {
		movies, err := m.fetchMovieFiltered(ctx, filter)
		if err != nil {
			return []models.Movie{}, err
		}
		return movies, nil
	}
}

func (m *MovieRepository) fetchMovieFiltered(ctx context.Context, filter models.MovieFilter) ([]models.Movie, error) {

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

	conds := []string{}     // accumulates SQL condition snippets (default: exclude archived)
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

// Movie Detail
func (m *MovieRepository) GetMovieDetails(ctx context.Context, movieId string) (models.Movie, error) {
	query := `
	select
		m.id,
		m.title,
		m.synopsis,
		m.poster_img,
		m.backdrop_img,
		m.duration_minutes,
		m.release_date,
		array_agg(
			distinct g.name
			order by
				g.name
		) as genres,

		-- Director
		d.name as director,

		-- Casts
		array_agg(
			distinct a.name
			order by
				a.name
		) as cast
	from
		movies m
		join movie_genres mg on m.id = mg.movie_id
		join genres g on g.id = mg.genre_id
		join people d on d.id = m.director_id
		left join movie_actors ma on ma.movie_id = m.id
		left join people a on a.id = ma.actor_id
	where
		m.id = $1
	group by
		m.id,
		m.title,
		m.synopsis,
		m.poster_img,
		m.backdrop_img,
		m.duration_minutes,
		m.release_date,
		d.name;
	`

	var movieDetails models.Movie
	err := m.db.QueryRow(ctx, query, movieId).Scan(
		&movieDetails.ID,
		&movieDetails.Title,
		&movieDetails.Synopsis,
		&movieDetails.PosterImg,
		&movieDetails.BackdropImg,
		&movieDetails.DurationMinutes,
		&movieDetails.ReleaseDate,
		&movieDetails.Genres,
		&movieDetails.Director,
		&movieDetails.Cast,
	)
	if err != nil {
		return models.Movie{}, err
	}

	return movieDetails, nil
}

// Archive a movie (delete)
func (m *MovieRepository) ArchiveMovieByID(ctx context.Context, movieId string) (models.ArchiveMovieRespond, error) {

	// Query
	query := `
	UPDATE movies
	SET 
		archived_at = CURRENT_TIMESTAMP,
		updated_at = CURRENT_TIMESTAMP
	WHERE id = $1 returning id, title, archived_at
	`

	var archivedMovie models.ArchiveMovieRespond
	err := m.db.QueryRow(ctx, query, movieId).Scan(&archivedMovie.ID, &archivedMovie.Title, &archivedMovie.Archived_at)
	if err != nil {
		return models.ArchiveMovieRespond{}, err
	}

	return archivedMovie, nil
}

func (m *MovieRepository) ListAllMovies(ctx context.Context) ([]models.Movie, error) {
	// Query for getting movies list (admin)
	query := `
	SELECT
		m.id,
		m.title,
		m.poster_img,
		m.release_date,
		ARRAY_AGG(DISTINCT g.name ORDER BY g.name) AS genres,
		m.duration_minutes
	FROM
		movies m
		JOIN movie_genres mg ON m.id = mg.movie_id
		JOIN genres g ON mg.genre_id = g.id
	GROUP BY
		m.id,
		m.title,
		m.poster_img,
		m.release_date,
		m.duration_minutes
	ORDER BY m.release_date DESC
	LIMIT 10 OFFSET 0;`
	rows, err := m.db.Query(ctx, query)
	if err != nil {
		log.Println("internal server error : ", err.Error())
		return []models.Movie{}, err
	}
	defer rows.Close()

	var movies []models.Movie

	// Read rows/records
	for rows.Next() {
		var movie models.Movie
		if err := rows.Scan(
			&movie.ID,
			&movie.Title,
			&movie.PosterImg,
			&movie.ReleaseDate,
			&movie.Genres,
			&movie.DurationMinutes,
		); err != nil {
			log.Println("scan error, ", err.Error())
			return []models.Movie{}, err
		}
		movies = append(movies, movie)
	}

	return movies, nil
}

func (m *MovieRepository) CreateMovie(ctx context.Context, movie models.CreateMovie) (models.CreateMovie, error) {
	// Begin transaction
	tx, err := m.db.Begin(ctx)
	if err != nil {
		return models.CreateMovie{}, err
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				log.Println("failed to rollback transaction: ", rollbackErr)
			}
		}
	}()

	// --- --- Step 1: Insert Genres and get their IDs
	var insertedGenreIDs []int
	if len(movie.Genres) > 0 {
		insertedGenreIDs, err = m.insertGenres(ctx, tx, movie.Genres)
		if err != nil {
			return models.CreateMovie{}, err
		}
	}

	log.Println("inserted genre IDs : ", insertedGenreIDs)

	// --- --- Step 2: Actors and Director into People
	var insertedCastIDs []int
	if len(movie.Cast) > 0 {
		insertedCastIDs, err = m.insertPeople(ctx, tx, movie.Cast)
		if err != nil {
			return models.CreateMovie{}, err
		}
	}
	log.Println("inserted cast IDs : ", insertedCastIDs)

	var insertedDirectorID int
	if len(movie.Director) > 0 {
		insertedDirectorID, err = m.insertDirector(ctx, tx, movie.Director)
		if err != nil {
			return models.CreateMovie{}, err
		}
	}
	log.Println("inserted director ID : ", insertedDirectorID)

	// --- --- Step 3: Create a new movie
	var newMovieID int
	newMovieID, err = m.insertMovie(ctx, tx, movie, insertedDirectorID)
	log.Println("new movie ID : ", newMovieID)

	// --- --- Step 4: Insert Movie_Genres
	if len(insertedGenreIDs) > 0 {
		err = m.insertMovieGenres(ctx, tx, newMovieID, insertedGenreIDs)
		if err != nil {
			return models.CreateMovie{}, err
		}
	}

	// --- --- Step 5: Insert Movie_Actors
	if len(insertedCastIDs) > 0 {
		err = m.insertMovieActors(ctx, tx, newMovieID, insertedCastIDs)
		if err != nil {
			return models.CreateMovie{}, err
		}
	}

	// (BELUM) Masukkan ke schedule

	// Commit transaction if everything succeeds
	if err = tx.Commit(ctx); err != nil {
		return models.CreateMovie{}, err
	}

	return models.CreateMovie{ID: newMovieID}, nil
}

func (m *MovieRepository) insertGenres(ctx context.Context, tx pgx.Tx, genres []string) ([]int, error) {
	if len(genres) == 0 {
		return []int{}, nil
	}

	var insertedGenreIDs []int

	for _, g := range genres {
		query := `
		WITH ins AS (
			INSERT INTO genres (name)
			VALUES ($1)
			ON CONFLICT (name) DO NOTHING
			RETURNING id
		)
		SELECT id FROM ins
		UNION
		SELECT id FROM genres WHERE name = $1;
		`

		var id int
		err := tx.QueryRow(ctx, query, g).Scan(&id)
		if err != nil {
			return nil, err
		}
		insertedGenreIDs = append(insertedGenreIDs, id)
	}

	return insertedGenreIDs, nil
}

func (m *MovieRepository) insertPeople(ctx context.Context, tx pgx.Tx, people []string) ([]int, error) {
	if len(people) == 0 {
		return []int{}, nil
	}

	var insertedPeopleIDs []int

	for _, g := range people {
		query := `
		WITH ins AS (
			INSERT INTO people (name)
			VALUES ($1)
			ON CONFLICT (name) DO NOTHING
			RETURNING id
		)
		SELECT id FROM ins
		UNION
		SELECT id FROM people WHERE name = $1;
		`

		var id int
		err := tx.QueryRow(ctx, query, g).Scan(&id)
		if err != nil {
			return nil, err
		}
		insertedPeopleIDs = append(insertedPeopleIDs, id)
	}

	return insertedPeopleIDs, nil
}

func (m *MovieRepository) insertDirector(ctx context.Context, tx pgx.Tx, director string) (int, error) {
	var insertedDirectorID int

	query := `
		WITH ins AS (
			INSERT INTO people (name)
			VALUES ($1)
			ON CONFLICT (name) DO NOTHING
			RETURNING id
		)
		SELECT id FROM ins
		UNION
		SELECT id FROM people WHERE name = $1;
		`

	err := tx.QueryRow(ctx, query, director).Scan(&insertedDirectorID)
	if err != nil {
		return 0, err
	}

	return insertedDirectorID, nil
}

func (m *MovieRepository) insertMovie(ctx context.Context, tx pgx.Tx, body models.CreateMovie, directorID int) (int, error) {
	var insertedMovieID int

	query := `
		insert into movies (title, synopsis, poster_img, backdrop_img, duration_minutes, release_date, director_id, age_rating_id) values ($1, $2, $3, $4, $5, $6, $7, $8) returning id`

	err := tx.QueryRow(ctx, query, body.Title, body.Synopsis, body.PosterImg, body.BackdropImg, body.DurationMinutes, body.ReleaseDate, directorID, body.AgeRating).Scan(&insertedMovieID)
	if err != nil {
		return 0, err
	}

	return insertedMovieID, nil
}

func (m *MovieRepository) insertMovieGenres(ctx context.Context, tx pgx.Tx, movieID int, genres []int) error {
	if len(genres) == 0 {
		return nil
	}

	// Build the bulk insert query
	query := `
		INSERT INTO movie_genres (movie_id, genre_id)
		SELECT $1, UNNEST($2::int[])
		ON CONFLICT DO NOTHING;
	`

	_, err := tx.Exec(ctx, query, movieID, genres)
	if err != nil {
		return fmt.Errorf("failed to insert movie_genres: %w", err)
	}

	return nil
}

func (m *MovieRepository) insertMovieActors(ctx context.Context, tx pgx.Tx, movieID int, actorIDs []int) error {
	if len(actorIDs) == 0 {
		return nil
	}

	query := `
		INSERT INTO movie_actors (movie_id, actor_id)
		SELECT $1, UNNEST($2::int[])
		ON CONFLICT DO NOTHING;
	`

	_, err := tx.Exec(ctx, query, movieID, actorIDs)
	if err != nil {
		return fmt.Errorf("failed to insert movie_actors: %w", err)
	}

	return nil
}
