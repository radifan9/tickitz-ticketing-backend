-- public.transactions_seats definition

-- Drop table

-- DROP TABLE public.transactions_seats;

CREATE TABLE public.transactions_seats (
	seats_id int4 NOT NULL,
	transactions_id uuid NOT NULL,
	CONSTRAINT transactions_seats_pkey PRIMARY KEY (seats_id, transactions_id)
);


-- public.transactions_seats foreign keys

ALTER TABLE public.transactions_seats ADD CONSTRAINT transactions_seats_seats_id_fkey FOREIGN KEY (seats_id) REFERENCES public.seat_codes(id);
ALTER TABLE public.transactions_seats ADD CONSTRAINT transactions_seats_transactions_id_fkey FOREIGN KEY (transactions_id) REFERENCES public.transactions(id);