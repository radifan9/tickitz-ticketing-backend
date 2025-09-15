-- public.users definition

-- Drop table

-- DROP TABLE public.users;

CREATE TABLE public.users (
	id uuid DEFAULT gen_random_uuid() NOT NULL,
	"role" public."role_type" DEFAULT 'user'::role_type NULL,
	email text NOT NULL,
	"password" text NOT NULL,
	created_at timestamptz DEFAULT CURRENT_TIMESTAMP NULL,
	updated_at timestamptz DEFAULT CURRENT_TIMESTAMP NULL,
	CONSTRAINT users_email_key UNIQUE (email),
	CONSTRAINT users_pkey PRIMARY KEY (id)
);