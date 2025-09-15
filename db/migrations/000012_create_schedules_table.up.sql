-- public.schedules definition

-- Drop table

-- DROP TABLE public.schedules;

CREATE TABLE public.schedules (
	id int4 GENERATED ALWAYS AS IDENTITY( INCREMENT BY 1 MINVALUE 1 MAXVALUE 2147483647 START 1 CACHE 1 NO CYCLE) NOT NULL,
	movie_id int4 NOT NULL,
	city_id int4 NOT NULL,
	show_time_id int4 NOT NULL,
	cinema_id int4 NOT NULL,
	show_date date NOT NULL,
	CONSTRAINT schedules_pkey PRIMARY KEY (id)
);


-- public.schedules foreign keys

ALTER TABLE public.schedules ADD CONSTRAINT schedules_cinema_id_fkey FOREIGN KEY (cinema_id) REFERENCES public.cinemas(id);
ALTER TABLE public.schedules ADD CONSTRAINT schedules_city_id_fkey FOREIGN KEY (city_id) REFERENCES public.cities(id);
ALTER TABLE public.schedules ADD CONSTRAINT schedules_movie_id_fkey FOREIGN KEY (movie_id) REFERENCES public.movies(id);
ALTER TABLE public.schedules ADD CONSTRAINT schedules_show_time_id_fkey FOREIGN KEY (show_time_id) REFERENCES public.show_times(id);