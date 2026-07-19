-- Primeira amostragem de permissions: squads:manage na app TALLO-PLANNING,
-- composta na role ADMIN. O tallo-api aplica o gate por permissão nas rotas de squads.
DO $$
DECLARE
    v_app UUID;
    v_perm UUID;
    v_role UUID;
BEGIN
    SELECT id INTO v_app FROM applications WHERE name = 'TALLO-PLANNING';
    IF v_app IS NULL THEN RETURN; END IF;

    INSERT INTO permissions (application_id, name)
    VALUES (v_app, 'squads:manage')
    ON CONFLICT (application_id, name) DO NOTHING;

    SELECT id INTO v_perm FROM permissions WHERE application_id = v_app AND name = 'squads:manage';
    SELECT id INTO v_role FROM roles WHERE application_id = v_app AND name = 'ADMIN';

    IF v_role IS NOT NULL THEN
        INSERT INTO role_permissions (role_id, permission_id)
        VALUES (v_role, v_perm)
        ON CONFLICT DO NOTHING;
    END IF;
END $$;
