-- public.show_times definition

-- Drop table

-- DROP TABLE public.show_times;

CREATE TABLE public.show_times (
	id int4 GENERATED ALWAYS AS IDENTITY( INCREMENT BY 1 MINVALUE 1 MAXVALUE 2147483647 START 1 CACHE 1 NO CYCLE) NOT NULL,
	start_at time NOT NULL,
	CONSTRAINT show_times_pkey PRIMARY KEY (id),
	CONSTRAINT show_times_start_at_key UNIQUE (start_at)
);