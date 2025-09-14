-- public.genres definition

-- Drop table

-- DROP TABLE public.genres;

CREATE TABLE public.genres (
	id int4 GENERATED ALWAYS AS IDENTITY( INCREMENT BY 1 MINVALUE 1 MAXVALUE 2147483647 START 1 CACHE 1 NO CYCLE) NOT NULL,
	"name" varchar(20) NOT NULL,
	CONSTRAINT genres_name_key UNIQUE (name),
	CONSTRAINT genres_pkey PRIMARY KEY (id)
);