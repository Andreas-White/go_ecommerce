-- Modify auths table to cascade delete on user deletion
ALTER TABLE public.auths
DROP CONSTRAINT IF EXISTS fk_user_id;

ALTER TABLE public.auths
ADD CONSTRAINT fk_user_id
FOREIGN KEY (user_id)
REFERENCES public.users (id)
ON UPDATE NO ACTION
ON DELETE CASCADE;

-- Modify products table to cascade delete on user deletion
ALTER TABLE public.products
DROP CONSTRAINT IF EXISTS fk_users_products;

ALTER TABLE public.products
ADD CONSTRAINT fk_users_products
FOREIGN KEY (user_id)
REFERENCES public.users (id)
ON UPDATE NO ACTION
ON DELETE CASCADE;
