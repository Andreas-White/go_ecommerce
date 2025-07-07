-- Add a nullable session_id column to the carts table
ALTER TABLE carts ADD COLUMN session_id VARCHAR(255);

-- Make the user_id column nullable
ALTER TABLE carts ALTER COLUMN user_id DROP NOT NULL;

-- Remove the existing unique constraint on user_id
-- The name of the constraint might vary depending on how it was created.
-- You may need to find the actual constraint name from your database schema.
-- This is a common default name: carts_user_id_key
-- If this fails, please replace 'carts_user_id_key' with the correct constraint name.
ALTER TABLE carts DROP CONSTRAINT IF EXISTS carts_user_id_key;

-- Add a partial unique index for user_id (only for non-null values)
CREATE UNIQUE INDEX idx_carts_user_id_not_null ON carts(user_id) WHERE user_id IS NOT NULL;

-- Add a partial unique index for session_id (only for non-null values)
CREATE UNIQUE INDEX idx_carts_session_id_not_null ON carts(session_id) WHERE session_id IS NOT NULL;

-- Add a check constraint to ensure that either user_id or session_id is set, but not both
ALTER TABLE carts ADD CONSTRAINT chk_user_or_session CHECK (
    (user_id IS NOT NULL AND session_id IS NULL) OR 
    (user_id IS NULL AND session_id IS NOT NULL)
);
