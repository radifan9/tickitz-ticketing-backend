-- public.movie_actors definition

-- Drop table

-- DROP TABLE public.movie_actors;

CREATE TABLE public.movie_actors (
	actor_id int4 NOT NULL,
	movie_id int4 NOT NULL,
	CONSTRAINT movie_actors_pkey PRIMARY KEY (actor_id, movie_id)
);


-- public.movie_actors foreign keys

ALTER TABLE public.movie_actors ADD CONSTRAINT movie_actors_actor_id_fkey FOREIGN KEY (actor_id) REFERENCES public.people(id);
ALTER TABLE public.movie_actors ADD CONSTRAINT movie_actors_movie_id_fkey FOREIGN KEY (movie_id) REFERENCES public.movies(id);