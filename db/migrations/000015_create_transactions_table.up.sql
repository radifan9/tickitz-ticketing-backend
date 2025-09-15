-- public.transactions definition

-- Drop table

-- DROP TABLE public.transactions;

CREATE TABLE public.transactions (
	id uuid DEFAULT gen_random_uuid() NOT NULL,
	user_id uuid NOT NULL,
	payment_id int4 NOT NULL,
	total_payment int4 NULL,
	full_name varchar(60) NULL,
	email text NULL,
	phone_number varchar(13) NULL,
	paid_at timestamptz NULL,
	created_at timestamptz DEFAULT CURRENT_TIMESTAMP NULL,
	updated_at timestamptz DEFAULT CURRENT_TIMESTAMP NULL,
	schedule_id int4 NULL,
	scanned_at timestamptz NULL,
	CONSTRAINT transactions_pkey PRIMARY KEY (id)
);


-- public.transactions foreign keys

ALTER TABLE public.transactions ADD CONSTRAINT transactions_payment_id_fkey FOREIGN KEY (payment_id) REFERENCES public.payments(id);
ALTER TABLE public.transactions ADD CONSTRAINT transactions_schedule_id_fkey FOREIGN KEY (schedule_id) REFERENCES public.schedules(id);
ALTER TABLE public.transactions ADD CONSTRAINT transactions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);