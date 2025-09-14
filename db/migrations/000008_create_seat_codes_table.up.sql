-- public.seat_codes definition

-- Drop table

-- DROP TABLE public.seat_codes;

CREATE TABLE public.seat_codes (
	id int4 GENERATED ALWAYS AS IDENTITY( INCREMENT BY 1 MINVALUE 1 MAXVALUE 2147483647 START 1 CACHE 1 NO CYCLE) NOT NULL,
	seat_code varchar(3) NOT NULL,
	CONSTRAINT seat_codes_pkey PRIMARY KEY (id)
);