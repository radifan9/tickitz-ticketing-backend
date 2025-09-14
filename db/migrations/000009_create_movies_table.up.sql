-- public.movies definition

-- Drop table

-- DROP TABLE public.movies;

CREATE TABLE public.movies (
	id int4 GENERATED ALWAYS AS IDENTITY( INCREMENT BY 1 MINVALUE 1 MAXVALUE 2147483647 START 1 CACHE 1 NO CYCLE) NOT NULL,
	title text NOT NULL,
	synopsis text NULL,
	poster_img text NULL,
	backdrop_img text NULL,
	duration_minutes int4 NULL,
	release_date date NULL,
	director_id int4 NULL,
	age_rating_id int4 NULL,
	created_at timestamptz DEFAULT CURRENT_TIMESTAMP NULL,
	updated_at timestamptz DEFAULT CURRENT_TIMESTAMP NULL,
	archived_at timestamptz NULL,
	CONSTRAINT movies_pkey PRIMARY KEY (id)
);


-- public.movies foreign keys

ALTER TABLE public.movies ADD CONSTRAINT fk_age_rating FOREIGN KEY (age_rating_id) REFERENCES public.age_ratings(id);
ALTER TABLE public.movies ADD CONSTRAINT movies_director_id_fkey FOREIGN KEY (director_id) REFERENCES public.people(id);