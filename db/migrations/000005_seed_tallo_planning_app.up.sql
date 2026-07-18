-- Cria a aplicação TALLO-PLANNING com as roles do domínio (ADMIN e USER).
-- A aplicação exige pessoa vinculada ao usuário para autenticar (require_person).
DO $$
DECLARE
    v_app_id UUID;
BEGIN
    INSERT INTO applications (name, auth_methods, require_person)
    VALUES ('TALLO-PLANNING', '{"PASSWORD"}', TRUE)
    RETURNING id INTO v_app_id;

    INSERT INTO roles (application_id, name) VALUES (v_app_id, 'ADMIN');
    INSERT INTO roles (application_id, name) VALUES (v_app_id, 'USER');
END $$;
