ALTER TABLE applications DROP COLUMN IF EXISTS require_person;

DROP INDEX IF EXISTS uq_persons_user_id;
