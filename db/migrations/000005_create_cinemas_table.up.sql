-- public.cinemas definition

-- Drop table

-- DROP TABLE public.cinemas;

CREATE TABLE public.cinemas (
	id int4 GENERATED ALWAYS AS IDENTITY( INCREMENT BY 1 MINVALUE 1 MAXVALUE 2147483647 START 1 CACHE 1 NO CYCLE) NOT NULL,
	"name" text NOT NULL,
	img text NOT NULL,
	ticket_price int4 NOT NULL,
	CONSTRAINT cinemas_pkey PRIMARY KEY (id)
);