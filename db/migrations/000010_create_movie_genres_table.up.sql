-- public.movie_genres definition

-- Drop table

-- DROP TABLE public.movie_genres;

CREATE TABLE public.movie_genres (
	movie_id int4 NOT NULL,
	genre_id int4 NOT NULL,
	CONSTRAINT movie_genres_pkey PRIMARY KEY (movie_id, genre_id)
);


-- public.movie_genres foreign keys

ALTER TABLE public.movie_genres ADD CONSTRAINT movie_genres_genre_id_fkey FOREIGN KEY (genre_id) REFERENCES public.genres(id);
ALTER TABLE public.movie_genres ADD CONSTRAINT movie_genres_movie_id_fkey FOREIGN KEY (movie_id) REFERENCES public.movies(id);