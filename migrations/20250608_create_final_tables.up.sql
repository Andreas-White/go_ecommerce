-- Create companies table
CREATE TABLE IF NOT EXISTS public.companies
(
    id uuid NOT NULL,
    user_id uuid NOT NULL,
    name text COLLATE pg_catalog."default" NOT NULL,
    address text COLLATE pg_catalog."default",
    city text COLLATE pg_catalog."default",
    country text COLLATE pg_catalog."default",
    zip_code text COLLATE pg_catalog."default",
    review_average numeric(3, 2) DEFAULT 0.00,
    review_count integer DEFAULT 0,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone,
    CONSTRAINT companies_pkey PRIMARY KEY (id),
    CONSTRAINT fk_company_user FOREIGN KEY (user_id)
        REFERENCES public.users (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE CASCADE 
);

ALTER TABLE IF EXISTS public.companies
    OWNER to postgres;

-- Create orders table
CREATE TABLE IF NOT EXISTS public.orders
(
    id uuid NOT NULL,
    user_id uuid NOT NULL,
    total_amount numeric NOT NULL CHECK (total_amount >= 0),
    status text COLLATE pg_catalog."default" NOT NULL DEFAULT 'pending', -- e.g., pending, processing, shipped, delivered, cancelled
    payment_status text COLLATE pg_catalog."default" NOT NULL DEFAULT 'pending', -- e.g., pending, paid, failed
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone,
    CONSTRAINT orders_pkey PRIMARY KEY (id),
    CONSTRAINT fk_order_user FOREIGN KEY (user_id)
        REFERENCES public.users (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE RESTRICT 
);

ALTER TABLE IF EXISTS public.orders
    OWNER to postgres;

-- Create order_items table
CREATE TABLE IF NOT EXISTS public.order_items
(
    id uuid NOT NULL,
    order_id uuid NOT NULL,
    product_id uuid NOT NULL,
    quantity integer NOT NULL CHECK (quantity > 0),
    price numeric NOT NULL CHECK (price >= 0), 
    CONSTRAINT order_items_pkey PRIMARY KEY (id),
    CONSTRAINT fk_order_item_order FOREIGN KEY (order_id)
        REFERENCES public.orders (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE CASCADE, 
    CONSTRAINT fk_order_item_product FOREIGN KEY (product_id)
        REFERENCES public.products (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE RESTRICT 
);

ALTER TABLE IF EXISTS public.order_items
    OWNER to postgres;

-- Create carts table
CREATE TABLE IF NOT EXISTS public.carts
(
    id uuid NOT NULL,
    user_id uuid NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone,
    CONSTRAINT carts_pkey PRIMARY KEY (id),
    CONSTRAINT fk_cart_user FOREIGN KEY (user_id)
        REFERENCES public.users (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE CASCADE, 
    CONSTRAINT unique_user_cart UNIQUE (user_id) 
);

ALTER TABLE IF EXISTS public.carts
    OWNER to postgres;

-- Create cart_items table
CREATE TABLE IF NOT EXISTS public.cart_items
(
    id uuid NOT NULL,
    cart_id uuid NOT NULL,
    product_id uuid NOT NULL,
    quantity integer NOT NULL CHECK (quantity > 0),
    price numeric NOT NULL CHECK (price >= 0),  
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT cart_items_pkey PRIMARY KEY (id),
    CONSTRAINT fk_cart_item_cart FOREIGN KEY (cart_id)
        REFERENCES public.carts (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE CASCADE, 
    CONSTRAINT fk_cart_item_product FOREIGN KEY (product_id)
        REFERENCES public.products (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE CASCADE, 
    CONSTRAINT unique_cart_product UNIQUE (cart_id, product_id)
);

ALTER TABLE IF EXISTS public.cart_items
    OWNER to postgres;

-- Create reviews table
CREATE TABLE IF NOT EXISTS public.reviews
(
    id uuid NOT NULL,
    product_id uuid NOT NULL,
    user_id uuid NOT NULL,
    rating integer NOT NULL CHECK (rating >= 1 AND rating <= 5),
    comment text COLLATE pg_catalog."default",
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT reviews_pkey PRIMARY KEY (id),
    CONSTRAINT fk_review_product FOREIGN KEY (product_id)
        REFERENCES public.products (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE CASCADE, 
    CONSTRAINT fk_review_user FOREIGN KEY (user_id)
        REFERENCES public.users (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE CASCADE, 
    CONSTRAINT unique_user_product_review UNIQUE (user_id, product_id)
);

ALTER TABLE IF EXISTS public.reviews
    OWNER to postgres;

-- Create payments table
CREATE TABLE IF NOT EXISTS public.payments
(
    id uuid NOT NULL,
    order_id uuid NOT NULL,
    amount numeric NOT NULL CHECK (amount >= 0),
    payment_method text COLLATE pg_catalog."default" NOT NULL, 
    status text COLLATE pg_catalog."default" NOT NULL DEFAULT 'pending', 
    transaction_id text COLLATE pg_catalog."default" UNIQUE,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT payments_pkey PRIMARY KEY (id),
    CONSTRAINT fk_payment_order FOREIGN KEY (order_id)
        REFERENCES public.orders (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE RESTRICT, 
    CONSTRAINT unique_payment_order UNIQUE (order_id) 
);

ALTER TABLE IF EXISTS public.payments
    OWNER to postgres;

-- Create shippings table
CREATE TABLE IF NOT EXISTS public.shippings
(
    id uuid NOT NULL,
    order_id uuid NOT NULL,
    method text COLLATE pg_catalog."default",
    tracking_code text COLLATE pg_catalog."default",
    cost numeric CHECK (cost >= 0),
    address text COLLATE pg_catalog."default" NOT NULL,
    city text COLLATE pg_catalog."default" NOT NULL,
    country text COLLATE pg_catalog."default" NOT NULL,
    zip_code text COLLATE pg_catalog."default" NOT NULL,
    shipped_at timestamp without time zone,
    delivered_at timestamp without time zone,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone,
    CONSTRAINT shippings_pkey PRIMARY KEY (id),
    CONSTRAINT fk_shipping_order FOREIGN KEY (order_id)
        REFERENCES public.orders (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE CASCADE, 
    CONSTRAINT unique_shipping_order UNIQUE (order_id) 
);

ALTER TABLE IF EXISTS public.shippings
    OWNER to postgres;