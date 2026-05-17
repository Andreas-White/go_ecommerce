DROP TABLE IF EXISTS user_auth_providers;
DROP TABLE IF EXISTS email_verification_tokens;
ALTER TABLE auths ALTER COLUMN password SET NOT NULL;
ALTER TABLE users DROP COLUMN IF EXISTS is_email_verified;