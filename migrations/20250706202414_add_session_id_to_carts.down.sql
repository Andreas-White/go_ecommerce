-- Before running this down migration, you should ensure that any carts
-- with a NULL user_id (i.e., guest carts) are either deleted or migrated
-- to a user account. Otherwise, the NOT NULL constraint will fail.
-- Example cleanup (use with caution):
-- DELETE FROM carts WHERE user_id IS NULL;

-- Remove the check constraint
ALTER TABLE carts DROP CONSTRAINT IF EXISTS chk_user_or_session;

-- Remove the partial unique indexes
DROP INDEX IF EXISTS idx_carts_user_id_not_null;
DROP INDEX IF EXISTS idx_carts_session_id_not_null;

-- Remove the session_id column
ALTER TABLE carts DROP COLUMN session_id;

-- Make the user_id column NOT NULL again
-- This will fail if there are any carts with a NULL user_id
ALTER TABLE carts ALTER COLUMN user_id SET NOT NULL;

-- Re-add the unique constraint on user_id
ALTER TABLE carts ADD CONSTRAINT carts_user_id_key UNIQUE (user_id);
