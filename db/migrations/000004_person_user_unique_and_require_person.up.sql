-- Garante a relação 1:1 entre usuário e pessoa.
-- Índice único parcial: múltiplas pessoas sem usuário (user_id NULL) continuam permitidas.
CREATE UNIQUE INDEX IF NOT EXISTS uq_persons_user_id ON persons (user_id) WHERE user_id IS NOT NULL;

-- Aplicações podem exigir que o usuário tenha uma pessoa vinculada para autenticar.
ALTER TABLE applications ADD COLUMN IF NOT EXISTS require_person BOOLEAN NOT NULL DEFAULT FALSE;
