-- public.people definition

-- Drop table

-- DROP TABLE public.people;

CREATE TABLE public.people (
	id int4 GENERATED ALWAYS AS IDENTITY( INCREMENT BY 1 MINVALUE 1 MAXVALUE 2147483647 START 1 CACHE 1 NO CYCLE) NOT NULL,
	"name" text NOT NULL,
	CONSTRAINT people_name_key UNIQUE (name),
	CONSTRAINT people_pkey PRIMARY KEY (id)
);