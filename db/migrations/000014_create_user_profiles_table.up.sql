-- public.user_profiles definition

-- Drop table

-- DROP TABLE public.user_profiles;

CREATE TABLE public.user_profiles (
	user_id uuid NOT NULL,
	first_name varchar(30) NULL,
	last_name varchar(30) NULL,
	img text NULL,
	phone_number varchar(13) NULL,
	points int4 NULL,
	created_at timestamptz DEFAULT CURRENT_TIMESTAMP NULL,
	updated_at timestamptz DEFAULT CURRENT_TIMESTAMP NULL,
	CONSTRAINT user_profiles_pkey PRIMARY KEY (user_id)
);


-- public.user_profiles foreign keys

ALTER TABLE public.user_profiles ADD CONSTRAINT user_profiles_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);