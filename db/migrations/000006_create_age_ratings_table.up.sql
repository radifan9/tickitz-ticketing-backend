-- public.age_ratings definition

-- Drop table

-- DROP TABLE public.age_ratings;

CREATE TABLE public.age_ratings (
	id int4 GENERATED ALWAYS AS IDENTITY( INCREMENT BY 1 MINVALUE 1 MAXVALUE 2147483647 START 1 CACHE 1 NO CYCLE) NOT NULL,
	age_rating text NOT NULL,
	CONSTRAINT age_ratings_age_rating_key UNIQUE (age_rating),
	CONSTRAINT age_ratings_pkey PRIMARY KEY (id)
);