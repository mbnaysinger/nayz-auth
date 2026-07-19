-- Role USER-FULL da app TALLO-PLANNING: poderes de ADMIN restritos à própria raia
-- (clonar, excluir, editar tipo/projeto, mover fora da semana), sem visão de terceiros.
INSERT INTO roles (application_id, name)
SELECT id, 'USER-FULL' FROM applications WHERE name = 'TALLO-PLANNING'
ON CONFLICT DO NOTHING;
