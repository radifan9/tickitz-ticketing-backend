package repositories

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"strconv"
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

// List Upcoming Movies (not yet released current_date < release_dateI)
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

// Fetching the Upcoming Movies
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
			AND m.archived_at IS NULL
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
			-- Make sure the movie is not archived

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
		AND m.archived_at IS NULL
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
	if filter.Limit == 20 && filter.Offset == 0 && len(filter.Keywords) == 0 && len(filter.Genres) == 0 {

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
		m.age_rating_id,
		array_agg(
			distinct g.name
			order by
				g.name
		) as genres,

		d.name as director,

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
		&movieDetails.AgeRatingID,
		&movieDetails.Genres,
		&movieDetails.Director,
		&movieDetails.Cast,
	)
	if err != nil {
		return models.Movie{}, err
	}

	return movieDetails, nil
}

// (admin) Archive a movie (delete)
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

	keysToInvalidate := []string{
		"tickitz:upcoming",
		"tickitz:popular",
		"tickitz:movies-all-first-page",
	}
	for _, k := range keysToInvalidate {
		if delErr := m.rdb.Del(ctx, k).Err(); delErr != nil {
			log.Printf("failed to invalidate cache for key %s: %v", k, delErr)
		}
	}

	return archivedMovie, nil
}

// (admin)
func (m *MovieRepository) ListAllMovies(ctx context.Context, offset int) ([]models.Movie, error) {
	// Query for getting movies list (admin)
	query := `
	SELECT
		m.id,
		m.title,
		m.poster_img,
		m.release_date,
		ARRAY_AGG(DISTINCT g.name ORDER BY g.name) AS genres,
		m.duration_minutes,
		m.created_at,
		m.updated_at
	FROM
		movies m
		JOIN movie_genres mg ON m.id = mg.movie_id
		JOIN genres g ON mg.genre_id = g.id
	WHERE
		m.archived_at IS NULL
	GROUP BY
		m.id,
		m.title,
		m.poster_img,
		m.release_date,
		m.duration_minutes,
		m.created_at,
		m.updated_at
	ORDER BY m.updated_at DESC
	LIMIT 10 OFFSET $1;`
	rows, err := m.db.Query(ctx, query, offset)
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
			&movie.CreatedAt,
			&movie.UpdatedAt,
		); err != nil {
			log.Println("scan error, ", err.Error())
			return []models.Movie{}, err
		}
		movies = append(movies, movie)
	}

	return movies, nil
}

// (admin)
func (m *MovieRepository) CreateMovie(ctx context.Context, movie models.CreateMovie, locationPoster string, locationBackdrop string) (models.CreateMovie, error) {
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

	// Step 1: Insert Genres and get their IDs
	var insertedGenreIDs []int
	genreList := strings.Split(movie.Genres, ",")
	if len(movie.Genres) > 0 {
		insertedGenreIDs, err = m.insertGenres(ctx, tx, genreList)
		if err != nil {
			log.Println("error inserting genres:", err)
			return models.CreateMovie{}, err
		}
	}
	log.Println("Inserted genre IDs : ", insertedGenreIDs)

	// Step 2: Actors and Director into People
	var insertedCastIDs []int
	castList := strings.Split(movie.Cast, ",")
	log.Println("cast : ", movie.Cast)
	if len(movie.Cast) > 0 {
		insertedCastIDs, err = m.insertPeople(ctx, tx, castList)
		if err != nil {
			log.Println("error inserting cast:", err)
			return models.CreateMovie{}, err
		}
	}
	log.Println("Inserted cast IDs : ", insertedCastIDs)

	var insertedDirectorID int
	if len(movie.Director) > 0 {
		insertedDirectorID, err = m.insertDirector(ctx, tx, movie.Director)
		if err != nil {
			log.Println("error inserting director:", err)
			return models.CreateMovie{}, err
		}
	}
	log.Println("Inserted director ID : ", insertedDirectorID)

	// Step 3: Create a new movie
	var newMovieID int
	newMovieID, err = m.insertMovie(ctx, tx, movie, insertedDirectorID, locationPoster, locationBackdrop)

	log.Println("new movie ID : ", newMovieID)

	// Step 4: Insert Movie_Genres
	if len(insertedGenreIDs) > 0 {
		err = m.insertMovieGenres(ctx, tx, newMovieID, insertedGenreIDs)
		if err != nil {
			log.Println("error inserting movie:", err)
			return models.CreateMovie{}, err
		}
	}

	// Step 5: Insert Movie_Actors
	if len(insertedCastIDs) > 0 {
		err = m.insertMovieActors(ctx, tx, newMovieID, insertedCastIDs)
		if err != nil {
			return models.CreateMovie{}, err
		}
	}

	// Step 6: Create schedules for 1 week from movie.ShowDate
	cinemaList := strings.Split(movie.CinemaID, ",")
	showTimeList := strings.Split(movie.ShowTimeID, ",")
	cityList := strings.Split(movie.CityID, ",")

	log.Println("Creating schedules...")
	log.Println("cinemaList : ", cinemaList)
	log.Println("show_date : ", movie.ShowDate)
	log.Println("show_time_ids : ", showTimeList)

	err = m.createSchedules(ctx, tx, newMovieID, movie.ShowDate, cinemaList, showTimeList, cityList)
	if err != nil {
		log.Println("error creating schedules:", err)
		return models.CreateMovie{}, err
	}

	// Commit transaction if everything succeeds
	if err = tx.Commit(ctx); err != nil {
		return models.CreateMovie{}, err
	}

	keysToInvalidate := []string{
		"tickitz:upcoming",
		"tickitz:popular",
		"tickitz:movies-all-first-page",
	}
	for _, k := range keysToInvalidate {
		if delErr := m.rdb.Del(ctx, k).Err(); delErr != nil {
			log.Printf("failed to invalidate cache for key %s: %v", k, delErr)
		}
	}

	return models.CreateMovie{ID: newMovieID, Title: movie.Title, ReleaseDate: movie.ReleaseDate}, nil
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

func (m *MovieRepository) insertMovie(ctx context.Context, tx pgx.Tx, body models.CreateMovie, directorID int, locationPoster string, locationBackdrop string) (int, error) {
	var insertedMovieID int

	query := `
		insert into
			movies (
				poster_img,
				backdrop_img,
				title,
				age_rating_id,
				release_date,
				duration_minutes,
				director_id,
				synopsis
			)
		values
			($1, $2, $3, $4, $5, $6, $7, $8) returning id`

	err := tx.QueryRow(ctx, query,
		locationPoster,
		locationBackdrop,
		body.Title,
		body.AgeRating,
		body.ReleaseDate,
		body.DurationMinutes,
		directorID,
		body.Synopsis).Scan(&insertedMovieID)
	if err != nil {
		log.Printf("insertMovie failed: %v", err)
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

// createSchedules creates schedules for a movie for 1 week starting from showDate
// for each combination of cinema, city, and show time
func (m *MovieRepository) createSchedules(ctx context.Context, tx pgx.Tx, movieID int, showDate string, cinemaIDs []string, showTimeIDs []string, cityIDs []string) error {
	if len(cinemaIDs) == 0 || len(showTimeIDs) == 0 || len(cityIDs) == 0 {
		log.Println("No cinemas, show times, or cities provided, skipping schedule creation")
		return nil
	}

	// Parse the show date
	startDate, err := time.Parse("2006-01-02", showDate)
	if err != nil {
		return fmt.Errorf("invalid show date format: %w", err)
	}

	// Convert cinema IDs from strings to integers
	var cinemaIntIDs []int
	for _, cidStr := range cinemaIDs {
		if cid, err := strconv.Atoi(cidStr); err == nil {
			cinemaIntIDs = append(cinemaIntIDs, cid)
		} else {
			log.Printf("Invalid cinema ID: %s, skipping", cidStr)
		}
	}

	// Convert show time IDs from strings to integers
	var showTimeIntIDs []int
	for _, stidStr := range showTimeIDs {
		if stid, err := strconv.Atoi(stidStr); err == nil {
			showTimeIntIDs = append(showTimeIntIDs, stid)
		} else {
			log.Printf("Invalid show time ID: %s, skipping", stidStr)
		}
	}

	// Convert city IDs from strings to integers
	var cityIntIDs []int
	for _, cityStr := range cityIDs {
		if cid, err := strconv.Atoi(cityStr); err == nil {
			cityIntIDs = append(cityIntIDs, cid)
		} else {
			log.Printf("Invalid city ID: %s, skipping", cityStr)
		}
	}

	if len(cinemaIntIDs) == 0 || len(showTimeIntIDs) == 0 || len(cityIntIDs) == 0 {
		return fmt.Errorf("no valid cinema IDs, show time IDs, or city IDs after parsing")
	}

	// Create schedules for 7 days (1 week)
	schedulesCreated := 0
	for day := 0; day < 7; day++ {
		currentDate := startDate.AddDate(0, 0, day)

		// For each city
		for _, cityID := range cityIntIDs {
			// For each cinema
			for _, cinemaID := range cinemaIntIDs {
				// For each show time
				for _, showTimeID := range showTimeIntIDs {
					scheduleQuery := `
						INSERT INTO schedules (movie_id, city_id, show_time_id, cinema_id, show_date)
						VALUES ($1, $2, $3, $4, $5)
					`

					_, err := tx.Exec(ctx, scheduleQuery,
						movieID,
						cityID,
						showTimeID,
						cinemaID,
						currentDate.Format("2006-01-02"))

					if err != nil {
						return fmt.Errorf("failed to insert schedule for movie %d, city %d, cinema %d, showtime %d, date %s: %w",
							movieID, cityID, cinemaID, showTimeID, currentDate.Format("2006-01-02"), err)
					}

					schedulesCreated++
					log.Printf("Created schedule: movie=%d, city=%d, cinema=%d, showtime=%d, date=%s",
						movieID, cityID, cinemaID, showTimeID, currentDate.Format("2006-01-02"))
				}
			}
		}
	}

	log.Printf("Successfully created %d schedules for movie %d", schedulesCreated, movieID)
	return nil
}

// (admin) EditMovie updates a movie and recreates its schedules for 7 days
func (m *MovieRepository) EditMovie(ctx context.Context, movieID int, movie models.CreateMovie, locationPoster string, locationBackdrop string) (models.CreateMovie, error) {
	tx, err := m.db.Begin(ctx)
	if err != nil {
		return models.CreateMovie{}, err
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				log.Println("failed to rollback transaction:", rollbackErr)
			}
		}
	}()

	// Step 1: Update movie info
	updateQuery := `
		UPDATE movies
		SET
			title = $1,
			age_rating_id = $2,
			release_date = $3,
			duration_minutes = $4,
			synopsis = $5,
			poster_img = COALESCE(NULLIF($6, ''), poster_img),
			backdrop_img = COALESCE(NULLIF($7, ''), backdrop_img),
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $8
		RETURNING id;
	`

	var updatedMovieID int
	err = tx.QueryRow(ctx, updateQuery,
		movie.Title,
		movie.AgeRating,
		movie.ReleaseDate,
		movie.DurationMinutes,
		movie.Synopsis,
		locationPoster,
		locationBackdrop,
		movieID,
	).Scan(&updatedMovieID)
	if err != nil {
		log.Printf("update movie failed: %v", err)
		return models.CreateMovie{}, err
	}

	// Step 2: Update director
	if len(movie.Director) > 0 {
		insertedDirectorID, err := m.insertDirector(ctx, tx, movie.Director)
		if err != nil {
			return models.CreateMovie{}, err
		}
		_, err = tx.Exec(ctx, `UPDATE movies SET director_id = $1 WHERE id = $2`, insertedDirectorID, movieID)
		if err != nil {
			return models.CreateMovie{}, err
		}
	}

	// Step 3: Update genres
	if len(movie.Genres) > 0 {
		genreList := strings.Split(movie.Genres, ",")
		insertedGenreIDs, err := m.insertGenres(ctx, tx, genreList)
		if err != nil {
			return models.CreateMovie{}, err
		}
		_, err = tx.Exec(ctx, `DELETE FROM movie_genres WHERE movie_id = $1`, movieID)
		if err != nil {
			return models.CreateMovie{}, err
		}
		err = m.insertMovieGenres(ctx, tx, movieID, insertedGenreIDs)
		if err != nil {
			return models.CreateMovie{}, err
		}
	}

	// Step 4: Update cast
	if len(movie.Cast) > 0 {
		castList := strings.Split(movie.Cast, ",")
		insertedCastIDs, err := m.insertPeople(ctx, tx, castList)
		if err != nil {
			return models.CreateMovie{}, err
		}
		_, err = tx.Exec(ctx, `DELETE FROM movie_actors WHERE movie_id = $1`, movieID)
		if err != nil {
			return models.CreateMovie{}, err
		}
		err = m.insertMovieActors(ctx, tx, movieID, insertedCastIDs)
		if err != nil {
			return models.CreateMovie{}, err
		}
	}

	// Step 5: Delete old schedules
	_, err = tx.Exec(ctx, `DELETE FROM schedules WHERE movie_id = $1`, movieID)
	if err != nil {
		log.Printf("failed to delete old schedules: %v", err)
		return models.CreateMovie{}, err
	}

	// Step 6: Create new schedules (7 days)
	cinemaList := strings.Split(movie.CinemaID, ",")
	showTimeList := strings.Split(movie.ShowTimeID, ",")
	cityList := strings.Split(movie.CityID, ",")

	err = m.createSchedules(ctx, tx, movieID, movie.ShowDate, cinemaList, showTimeList, cityList)
	if err != nil {
		log.Printf("error creating new schedules: %v", err)
		return models.CreateMovie{}, err
	}

	// Step 7: Commit
	if err = tx.Commit(ctx); err != nil {
		return models.CreateMovie{}, err
	}

	// Step 8: Invalidate caches
	keysToInvalidate := []string{
		"tickitz:upcoming",
		"tickitz:popular",
		"tickitz:movies-all-first-page",
	}
	for _, k := range keysToInvalidate {
		if delErr := m.rdb.Del(ctx, k).Err(); delErr != nil {
			log.Printf("failed to invalidate cache for key %s: %v", k, delErr)
		}
	}

	return models.CreateMovie{
		ID:              movieID,
		Title:           movie.Title,
		ReleaseDate:     movie.ReleaseDate,
		DurationMinutes: movie.DurationMinutes,
	}, nil
}
