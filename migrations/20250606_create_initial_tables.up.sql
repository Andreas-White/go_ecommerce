-- Create users table
CREATE TABLE IF NOT EXISTS public.users
(
    id uuid NOT NULL,
    first_name text COLLATE pg_catalog."default" NOT NULL,
    email text COLLATE pg_catalog."default" NOT NULL,
    phone numeric,
    last_name text COLLATE pg_catalog."default",
    middle_name text COLLATE pg_catalog."default",
    is_producer boolean DEFAULT false,
    address text COLLATE pg_catalog."default",
    city text COLLATE pg_catalog."default",
    country text COLLATE pg_catalog."default",
    zip_code numeric,
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    CONSTRAINT users_pkey PRIMARY KEY (id),
    CONSTRAINT unique_email UNIQUE (email)
);

ALTER TABLE IF EXISTS public.users
    OWNER to postgres;

-- Create auths table
CREATE TABLE IF NOT EXISTS public.auths
(
    id uuid NOT NULL,
    user_id uuid NOT NULL,
    created_at timestamp without time zone NOT NULL,
    active boolean NOT NULL,
    password text COLLATE pg_catalog."default",
    updated_at timestamp without time zone,
    CONSTRAINT auths_pkey PRIMARY KEY (id),
    CONSTRAINT fk_user_id FOREIGN KEY (user_id)
        REFERENCES public.users (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

ALTER TABLE IF EXISTS public.auths
    OWNER to postgres;

-- Create products table
CREATE TABLE IF NOT EXISTS public.products
(
    id uuid NOT NULL,
    name text COLLATE pg_catalog."default" NOT NULL,
    description text COLLATE pg_catalog."default",
    price numeric NOT NULL,
    stock numeric NOT NULL,
    category text COLLATE pg_catalog."default" NOT NULL,
    image_url text COLLATE pg_catalog."default",
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    user_id uuid NOT NULL,
    CONSTRAINT products_pkey PRIMARY KEY (id),
    CONSTRAINT fk_users_products FOREIGN KEY (user_id)
        REFERENCES public.users (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
        NOT VALID
);

ALTER TABLE IF EXISTS public.products
    OWNER to postgres;