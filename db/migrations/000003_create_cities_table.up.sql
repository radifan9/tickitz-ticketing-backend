-- public.cities definition

-- Drop table

-- DROP TABLE public.cities;

CREATE TABLE public.cities (
	id int4 GENERATED ALWAYS AS IDENTITY( INCREMENT BY 1 MINVALUE 1 MAXVALUE 2147483647 START 1 CACHE 1 NO CYCLE) NOT NULL,
	"name" text NOT NULL,
	CONSTRAINT cities_name_key UNIQUE (name),
	CONSTRAINT cities_pkey PRIMARY KEY (id)
);